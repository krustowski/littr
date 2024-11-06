package users

import (
	"context"
	"fmt"
	netmail "net/mail"
	"regexp"
	"strconv"
	"strings"
	"time"

	"go.vxn.dev/littr/pkg/backend/common"
	"go.vxn.dev/littr/pkg/backend/image"
	"go.vxn.dev/littr/pkg/backend/mail"
	//"go.vxn.dev/littr/pkg/backend/live"
	"go.vxn.dev/littr/pkg/backend/pages"
	"go.vxn.dev/littr/pkg/config"
	"go.vxn.dev/littr/pkg/helpers"
	"go.vxn.dev/littr/pkg/models"

	uuid "github.com/google/uuid"
)

type UserUpdateRequest struct {
	NewPassphraseHex     string `json:"new_passphrase_hex"`
	CurrentPassphraseHex string `json:"current_passphrase_hex"`
	AvatarByteData       []byte `json:"data"`
	AvatarFileName       string `json:"figure"`
}

//
// models.UserServiceInterface implementation
//

type UserService struct {
	postRepository    models.PostRepositoryInterface
	requestRepository models.RequestRepositoryInterface
	tokenRepository   models.TokenRepositoryInterface
	userRepository    models.UserRepositoryInterface
}

func NewUserService(
	postRepository models.PostRepositoryInterface,
	requestRepository models.RequestRepositoryInterface,
	tokenRepository models.TokenRepositoryInterface,
	userRepository models.UserRepositoryInterface,
) models.UserServiceInterface {
	if postRepository == nil || requestRepository == nil || tokenRepository == nil || userRepository == nil {
		return nil
	}

	return &UserService{
		postRepository:    postRepository,
		requestRepository: requestRepository,
		tokenRepository:   tokenRepository,
		userRepository:    userRepository,
	}
}

func (s *UserService) Create(ctx context.Context, user *models.User) error {
	// Check if the registration is allowed.
	if !config.IsRegistrationEnabled {
		return fmt.Errorf(common.ERR_REGISTRATION_DISABLED)
	}

	// Block restricted nicknames, use lowercase for comparison.
	if helpers.Contains(config.UserDeletionList, strings.ToLower(user.Nickname)) {
		return fmt.Errorf(common.ERR_RESTRICTED_NICKNAME)
	}

	// Restrict the nickname to contains only some characters explicitly "listed".
	// https://stackoverflow.com/a/38554480
	if !regexp.MustCompile(`^[a-zA-Z0-9]+$`).MatchString(user.Nickname) {
		return fmt.Errorf(common.ERR_NICKNAME_CHARSET_MISMATCH)
	}

	// Check the nick's length contraints.
	if len(user.Nickname) > 12 || len(user.Nickname) < 3 {
		return fmt.Errorf(common.ERR_NICKNAME_TOO_LONG_SHORT)
	}

	// Preprocess the e-mail address: set to lowercase.
	email := strings.ToLower(user.Email)
	user.Email = email

	// Validate the e-mail format.
	// https://stackoverflow.com/a/66624104
	if _, err := netmail.ParseAddress(email); err != nil {
		return fmt.Errorf(common.ERR_WRONG_EMAIL_FORMAT)
	}

	// Check if the e-mail address already used.
	users, err := s.userRepository.GetAll()
	if err != nil {
		return err
	}

	for key, dbUser := range *users {
		// Check if the nickname has been already used/taken.
		if key == user.Nickname {
			return fmt.Errorf(common.ERR_USER_NICKNAME_TAKEN)
		}

		// E-mail address match found.
		if strings.ToLower(dbUser.Email) == user.Email {
			return fmt.Errorf(common.ERR_EMAIL_ALREADY_USED)
		}
	}

	//
	//  Validation end
	//

	//
	//  Set user defaults, save the user struct to database and create a new system post
	//

	// Set the defaults and a timestamp.
	user.RegisteredTime = time.Now()
	user.LastActiveTime = time.Now()
	user.About = "newbie"

	// New user's umbrella option map.
	options := map[string]bool{
		"active":        false,
		"gdpr":          true,
		"private":       false,
		"uiDarkMode":    true,
		"liveMode":      true,
		"localTimeMode": true,
	}

	// Options defaults.
	user.Options = options

	// Deprecated option setting method.
	user.GDPR = true
	user.Active = false

	//
	//  Request (for activation) composition
	//

	// Generate new random UUID, aka requestID.
	randomID := uuid.New().String()

	// Prepare the request data for the database.
	request := &models.Request{
		ID:        randomID,
		Nickname:  user.Nickname,
		Email:     user.Email,
		CreatedAt: time.Now(),
		Type:      "activation",
	}

	// Save the request via the requestRepository.
	err = s.requestRepository.Save(request)
	if err != nil {
		return fmt.Errorf(common.ERR_REQUEST_SAVE_FAIL)
	}

	//
	//  Mailing
	//

	// Prepare the mail options.
	mailPayload := mail.MessagePayload{
		Nickname: user.Nickname,
		Email:    user.Email,
		Type:     "user_activation",
		UUID:     randomID,
	}

	// Compose a message to send.
	msg, err := mail.ComposeMail(mailPayload)
	if err != nil || msg == nil {
		return fmt.Errorf(common.ERR_MAIL_COMPOSITION_FAIL)
	}

	// Send the activation mail to such user.
	if err = mail.SendMail(msg); err != nil {
		return fmt.Errorf(common.ERR_ACTIVATION_MAIL_FAIL)
	}

	//
	//  Save new user
	//

	err = s.userRepository.Save(user)
	if err != nil {
		return fmt.Errorf(common.ERR_USER_SAVE_FAIL)
	}

	//
	//  Compose new post
	//

	// Prepare a timestamp for a very new post to alert others the new user has been added.
	postStamp := time.Now()
	postKey := strconv.FormatInt(postStamp.UnixNano(), 10)

	// Compose the post's body.
	post := &models.Post{
		ID:        postKey,
		Type:      "user",
		Figure:    user.Nickname,
		Nickname:  "system",
		Content:   "new user has been added (" + user.Nickname + ")",
		Timestamp: postStamp,
	}

	// Save new post to the database.
	err = s.postRepository.Save(post)
	if err != nil {
		return fmt.Errorf(common.ERR_POSTREG_POST_SAVE_FAIL)
	}

	return nil
}

func (s *UserService) Activate(ctx context.Context, UUID string) error {
	if UUID == "" {
		return fmt.Errorf(common.ERR_REQUEST_UUID_BLANK)
	}

	request, err := s.requestRepository.GetByID(UUID)
	if err != nil {
		// Hm, could be another error than just ERR_REQUEST_UUID_INVALID...
		//return fmt.Errorf(common.ERR_REQUEST_UUID_INVALID)
		return err
	}

	// Check the request's validity.
	if request.CreatedAt.Before(time.Now().Add(-24 * time.Hour)) {
		// Delete the expired request from database.
		err := s.requestRepository.Delete(UUID)
		if err != nil {
			return fmt.Errorf(common.ERR_REQUEST_DELETE_FAIL)
		}

		// Delete the expired inactivated user from database.
		err = s.userRepository.Delete(request.Nickname)
		if err != nil {
			return fmt.Errorf(common.ERR_USER_DELETE_FAIL)
		}

		return fmt.Errorf(common.ERR_REQUEST_UUID_EXPIRED)
	}

	// Fetch the request-related user from database.
	user, err := s.userRepository.GetByID(request.Nickname)
	if err != nil {
		return err
	}

	// Update the user's activation status (a deprecated and a new method).
	user.Active = true

	if user.Options == nil {
		user.Options = make(map[string]bool)
	}
	user.Options["active"] = true

	// Save the just-activated user back to the database.
	err = s.userRepository.Save(user)
	if err != nil {
		return fmt.Errorf(common.ERR_USER_UPDATE_FAIL)
	}

	// Delete the request from the request database.
	err = s.requestRepository.Delete(UUID)
	if err != nil {
		return fmt.Errorf(common.ERR_REQUEST_DELETE_FAIL)
	}

	return nil
}

func (s *UserService) Update(ctx context.Context, userRequest interface{}) error {
	// Assert the type for the user update request.
	_, ok := userRequest.(*UserUpdateRequest)
	if !ok {
		return fmt.Errorf("could not decode the user request")
	}

	// Fetch the update type from the given context.
	reqType, ok := ctx.Value("updateType").(string)
	if !ok {
		return fmt.Errorf("could not decode the user request")
	}

	switch reqType {
	case "lists":
	case "options":
	case "passhrase":
	default:
		return fmt.Errorf("unknown request type")
	}

	return fmt.Errorf("not yet implemented")
}

func (s *UserService) UpdateAvatar(ctx context.Context, userRequest interface{}) (*string, error) {
	// Assert the type for the user update request.
	data, ok := userRequest.(*UserUpdateRequest)
	if !ok {
		return nil, fmt.Errorf("could not decode the user request")
	}

	// Fetch the callerID/nickname type from the given context.
	callerID, ok := ctx.Value("nickname").(string)
	if !ok {
		return nil, fmt.Errorf("could not decode the user request")
	}

	// Fetch the user's data.
	user, err := s.userRepository.GetByID(callerID)
	if err != nil {
		return nil, err
	}

	imgData := &image.ImageProcessPayload{
		ImageByteData: &data.AvatarByteData,
		ImageFileName: data.AvatarFileName,
		ImageBaseName: fmt.Sprintf("%d", time.Now().UnixNano()),
	}

	// Uploaded figure handling.
	imageBaseURL, err := image.ProcessImageBytes(imgData)
	if err != nil {
		return nil, err
	}

	//
	//  TODO: remove previous avatar's data
	//

	user.AvatarURL = "/web/pix/thumb_" + *imageBaseURL

	// Update user's data.
	err = s.userRepository.Save(user)
	if err != nil {
		return nil, err
	}

	return imageBaseURL, nil
}

func (s *UserService) ProcessPassphraseRequest(ctx context.Context, userRequest interface{}) error {
	// Assert the type for the user update request.
	_, ok := userRequest.(*UserUpdateRequest)
	if !ok {
		return fmt.Errorf("could not decode the user request")
	}

	// Fetch the callerID/nickname type from the given context.
	_, ok = ctx.Value("nickname").(string)
	if !ok {
		return fmt.Errorf("could not decode the user request")
	}

	// Fetch the pageNo from the context.
	requestType, ok := ctx.Value("requestType").(string)
	if !ok {
		return fmt.Errorf(common.ERR_PAGENO_INCORRECT)
	}

	switch requestType {
	case "request":
	case "reset":
	default:
		return fmt.Errorf(common.ERR_REQUEST_TYPE_UNKNOWN)
	}

	return fmt.Errorf("not yet implemented")
}

func (s *UserService) Delete(ctx context.Context, userID string) error {
	return fmt.Errorf("not yet implemented")
}

func (s *UserService) FindAll(ctx context.Context) (*map[string]models.User, error) {
	// Fetch the caller's ID from the context.
	callerID, ok := ctx.Value("nickname").(string)
	if !ok {
		return nil, fmt.Errorf(common.ERR_CALLER_FAIL)
	}

	// Fetch the pageNo from the context.
	pageNo, ok := ctx.Value("pageNo").(int)
	if !ok {
		return nil, fmt.Errorf(common.ERR_PAGENO_INCORRECT)
	}

	// Request the caller from the user repository.
	caller, err := s.userRepository.GetByID(callerID)
	if err != nil {
		return nil, err
	}

	// Compose a pagination options object to paginate users.
	opts := &pages.PageOptions{
		CallerID: callerID,
		PageNo:   pageNo,
		FlowList: nil,

		Users: pages.UserOptions{
			Plain:       true,
			RequestList: &caller.RequestList,
		},
	}

	// Request the page of users from the user repository.
	users, err := s.userRepository.GetPage(opts)
	if err != nil {
		return nil, err
	}

	// Add the caller to users map.
	(*users)[callerID] = *caller

	// Patch the user's data for export.
	patchedUsers := common.FlushUserData(users, callerID)

	return patchedUsers, nil
}

func (s *UserService) FindByID(ctx context.Context, userID string) (*models.User, error) {
	// Fetch the user's ID from the context.
	userID, ok := ctx.Value("userID").(string)
	if !ok {
		return nil, fmt.Errorf(common.ERR_USERID_BLANK)
	}

	// Request the user's data from repository..
	user, err := s.userRepository.GetByID(userID)
	if err != nil {
		return nil, err
	}

	// Patch the user's data for export.
	patchedUser := (*common.FlushUserData(&map[string]models.User{userID: *user}, userID))[userID]

	return &patchedUser, nil
}
