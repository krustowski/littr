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
	// Lists update request payload.
	FlowList    map[string]bool `json:"flow_list"`
	RequestList map[string]bool `json:"request_list"`
	ShadeList   map[string]bool `json:"shade_list"`

	// Options updata request payload (legacy fields).
	UIDarkMode    bool                  `json:"dark_mode"`
	LiveMode      bool                  `json:"live_mode"`
	LocalTimeMode bool                  `json:"local_time_mode"`
	Private       bool                  `json:"private"`
	AboutText     string                `json:"about_you"`
	WebsiteLink   string                `json:"website_link"`
	OptionsMap    models.UserOptionsMap `json:"options_map"`

	// New passphrase request payload.
	NewPassphraseHex     string `json:"new_passphrase_hex"`
	CurrentPassphraseHex string `json:"current_passphrase_hex"`

	// New avatar upload/update request payload.
	AvatarByteData []byte `json:"data"`
	AvatarFileName string `json:"figure"`
}

//
// models.UserServiceInterface implementation
//

type UserService struct {
	postRepository         models.PostRepositoryInterface
	subscriptionRepository models.SubscriptionRepositoryInterface
	requestRepository      models.RequestRepositoryInterface
	tokenRepository        models.TokenRepositoryInterface
	userRepository         models.UserRepositoryInterface
}

func NewUserService(
	postRepository models.PostRepositoryInterface,
	subscriptionRepository models.SubscriptionRepositoryInterface,
	requestRepository models.RequestRepositoryInterface,
	tokenRepository models.TokenRepositoryInterface,
	userRepository models.UserRepositoryInterface,
) models.UserServiceInterface {
	if postRepository == nil || requestRepository == nil || tokenRepository == nil || userRepository == nil {
		return nil
	}

	return &UserService{
		postRepository:         postRepository,
		subscriptionRepository: subscriptionRepository,
		requestRepository:      requestRepository,
		tokenRepository:        tokenRepository,
		userRepository:         userRepository,
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
	data, ok := userRequest.(*UserUpdateRequest)
	if !ok {
		return fmt.Errorf("could not decode the user request")
	}

	// Fetch the callerID from the given context.
	callerID, ok := ctx.Value("nickname").(string)
	if !ok {
		return fmt.Errorf("could not decode the caller's ID")
	}

	// Fetch the userID from the given context.
	userID, ok := ctx.Value("userID").(string)
	if !ok {
		return fmt.Errorf("could not decode the user's ID")
	}

	// Fetch the update type from the given context.
	reqType, ok := ctx.Value("updateType").(string)
	if !ok {
		return fmt.Errorf("could not decode the user request type")
	}

	switch reqType {
	case "lists":
		// Fetch the caller from repository.
		caller, err := s.userRepository.GetByID(callerID)
		if err != nil {
			return fmt.Errorf(common.ERR_CALLER_NOT_FOUND)
		}

		// Fetch requested user by userID from repository.
		dbUser, err := s.userRepository.GetByID(userID)
		if err != nil {
			return fmt.Errorf(common.ERR_USER_NOT_FOUND)
		}

		// Process the flowList request.
		if data.FlowList != nil {
			dbUser = processFlowList(data, dbUser, caller, s.userRepository)
		}

		// Process the requestList request.
		if data.RequestList != nil {
			dbUser = processRequestList(data, dbUser, caller)
		}

		// Process the shadeList request.
		if data.ShadeList != nil {
			dbUser = processShadeList(data, dbUser, caller)
		}

		if err := s.userRepository.Save(dbUser); err != nil {
			return err
		}

	case "options":
		if callerID != userID {
			return fmt.Errorf(common.ERR_USER_UPDATE_FOREIGN)
		}

		// Fetch requested user by userID from repository.
		dbUser, err := s.userRepository.GetByID(userID)
		if err != nil {
			return fmt.Errorf(common.ERR_USER_NOT_FOUND)
		}
		// Patch the nil map.
		if dbUser.Options == nil {
			dbUser.Options = models.UserOptionsMap{}
		}

		// Toggle dark mode to light mode and vice versa.
		if data.UIDarkMode != dbUser.UIDarkMode {
			dbUser.UIDarkMode = !dbUser.UIDarkMode
			dbUser.Options["uiDarkMode"] = data.UIDarkMode
		}

		// Toggle the live mode.
		if data.LiveMode != dbUser.LiveMode {
			dbUser.LiveMode = !dbUser.LiveMode
			dbUser.Options["liveMode"] = data.LiveMode
		}

		// Toggle the local time mode.
		if data.LocalTimeMode != dbUser.LocalTimeMode {
			dbUser.LocalTimeMode = !dbUser.LocalTimeMode
			dbUser.Options["localTimeMode"] = data.LocalTimeMode
		}

		// Toggle the private mode.
		if data.Private != dbUser.Private {
			dbUser.Private = !dbUser.Private
			dbUser.Options["private"] = data.Private
		}

		// Change the about text if present and differs from the current one.
		if data.AboutText != "" && data.AboutText != dbUser.About {
			dbUser.About = data.AboutText
		}

		// Change the website link if present and differs from the current one.
		if data.WebsiteLink != "" && data.WebsiteLink != dbUser.Web {
			dbUser.Web = data.WebsiteLink
		}

		if err := s.userRepository.Save(dbUser); err != nil {
			return err
		}

	case "passphrase":
		if callerID != userID {
			return fmt.Errorf(common.ERR_USER_PASSPHRASE_FOREIGN)
		}

		// Fetch requested user from repository.
		dbUser, err := s.userRepository.GetByID(userID)
		if err != nil {
			return fmt.Errorf(common.ERR_USER_NOT_FOUND)
		}

		// Check if both new or old passphrase hashes are blank/empty.
		if data.NewPassphraseHex == "" || data.CurrentPassphraseHex == "" {
			return fmt.Errorf(common.ERR_PASSPHRASE_REQ_INCOMPLETE)
		}

		// Check if the current passphrasë́'s hash is correct.
		if data.CurrentPassphraseHex != dbUser.PassphraseHex {
			return fmt.Errorf(common.ERR_PASSPHRASE_CURRENT_WRONG)
		}

		// Update user's passphrase.
		dbUser.PassphraseHex = data.NewPassphraseHex

		if err := s.userRepository.Save(dbUser); err != nil {
			return err
		}

	default:
		return fmt.Errorf("unknown request type")
	}

	return nil
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
	// Fetch the caller's ID from the context.
	callerID, ok := ctx.Value("nickname").(string)
	if !ok {
		return fmt.Errorf(common.ERR_CALLER_FAIL)
	}

	// Fetch the user's ID from the context.
	userID, ok = ctx.Value("userID").(string)
	if !ok {
		return fmt.Errorf(common.ERR_USERID_BLANK)
	}

	// Check for possible user's data forgery attempt.
	if userID != callerID {
		return fmt.Errorf(common.ERR_USER_DELETE_FOREIGN)
	}

	// Fetch the user's data.
	_, err := s.userRepository.GetByID(userID)
	if err != nil {
		return fmt.Errorf(common.ERR_USER_NOT_FOUND)
	}

	// Delete requested user's record from database.
	if err := s.userRepository.Delete(userID); err != nil {
		return fmt.Errorf(common.ERR_USER_DELETE_FAIL)
	}

	// Delete requested user's subscription.
	if err := s.subscriptionRepository.Delete(userID); err != nil {
		return fmt.Errorf(common.ERR_SUBSCRIPTION_DELETE_FAIL)
	}

	//
	//  Delete all posts, delete polls, delete tokens
	//

	return nil
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

//
//  Helpers
//

func processFlowList(data *UserUpdateRequest, user *models.User, caller *models.User, r models.UserRepositoryInterface) *models.User {
	// Process the FlowList if not empty.
	if user.FlowList == nil {
		user.FlowList = make(map[string]bool)
	}

	// Loop over all flowList records.
	for key, value := range data.FlowList {
		// Forbid changing the foreign flowList according to the requested flowList records.
		if user.Nickname != caller.Nickname && key != caller.Nickname {
			continue
		}

		// Only allow to change controlling user's field in the foreign flowList.
		if user.Nickname != caller.Nickname && key == caller.Nickname {
			user.FlowList[key] = value
			continue
		}

		// Set such flowList record according to the request data.
		if _, found := user.FlowList[key]; found {
			user.FlowList[key] = value
		}

		// Check if the caller is shaded by the counterpart.
		counterpart, err := r.GetByID(key)
		if err == nil {
			if counterpart.Private && key != caller.Nickname {
				// cannot add this user to one's flow, as the following
				// has to be requested and allowed by the counterpart
				user.FlowList[key] = false
				continue
			}

			// update the flowList record according to the counterpart's shade list state of the user
			if shaded, found := counterpart.ShadeList[user.Nickname]; !found || (found && !shaded) {
				user.FlowList[key] = value
			}
		}

		// Do not allow to unfollow oneself.
		if key == user.Nickname {
			user.FlowList[key] = true
		}

	}
	// always allow to see system posts
	user.FlowList["system"] = true

	return user
}

func processRequestList(data *UserUpdateRequest, user *models.User, caller *models.User) *models.User {
	if user.RequestList == nil {
		user.RequestList = make(map[string]bool)
	}

	// Loop over the RequestList records and change the user's values accordingly (enforce the proper requestList changing!).
	for key, value := range data.RequestList {
		// Only allow to change the caller's record in the remote/counterpart's requestList.
		if key != caller.Nickname {
			continue
		}

		user.RequestList[key] = value
	}

	return user
}

func processShadeList(data *UserUpdateRequest, user *models.User, caller *models.User) *models.User {
	if user.ShadeList == nil {
		user.ShadeList = make(map[string]bool)
	}

	// Loop over the ShadeList records and change the user's values accordingly (enforce the proper shadeList changing!).
	for key, value := range data.ShadeList {
		// To change the shadeList, one has to be its owner.
		if user.Nickname != caller.Nickname {
			continue
		}

		user.ShadeList[key] = value
	}

	return user
}
