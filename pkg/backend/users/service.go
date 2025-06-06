package users

import (
	"context"
	"crypto/sha512"
	"fmt"
	"net/http"
	netmail "net/mail"
	"os"
	"reflect"
	"regexp"
	"slices"
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

//
// models.UserServiceInterface implementation
//

type UserService struct {
	mailService       models.MailServiceInterface
	pagingService     models.PagingServiceInterface
	pollRepository    models.PollRepositoryInterface
	postRepository    models.PostRepositoryInterface
	requestRepository models.RequestRepositoryInterface
	tokenRepository   models.TokenRepositoryInterface
	userRepository    models.UserRepositoryInterface
}

func NewUserService(
	mailService models.MailServiceInterface,
	pagingService models.PagingServiceInterface,
	pollRepository models.PollRepositoryInterface,
	postRepository models.PostRepositoryInterface,
	requestRepository models.RequestRepositoryInterface,
	tokenRepository models.TokenRepositoryInterface,
	userRepository models.UserRepositoryInterface,
) models.UserServiceInterface {

	if mailService == nil ||
		pagingService == nil ||
		pollRepository == nil ||
		postRepository == nil ||
		requestRepository == nil ||
		tokenRepository == nil ||
		userRepository == nil {

		return nil
	}

	return &UserService{
		mailService:       mailService,
		pagingService:     pagingService,
		pollRepository:    pollRepository,
		postRepository:    postRepository,
		requestRepository: requestRepository,
		tokenRepository:   tokenRepository,
		userRepository:    userRepository,
	}
}

func (s *UserService) Create(ctx context.Context, createRequestI interface{}) error {
	// Check if the registration is allowed.
	if !config.IsRegistrationEnabled {
		return fmt.Errorf(common.ERR_REGISTRATION_DISABLED)
	}

	createRequest, ok := createRequestI.(*UserCreateRequest)
	if !ok {
		return fmt.Errorf("invalid data inserted, cannot continue processing the request")
	}

	// Block restricted nicknames, use lowercase for comparison.
	if helpers.Contains(config.UserDeletionList, strings.ToLower(createRequest.Nickname)) {
		return fmt.Errorf(common.ERR_RESTRICTED_NICKNAME)
	}

	// Restrict the nickname to contains only some characters explicitly "listed".
	// https://stackoverflow.com/a/38554480
	if !regexp.MustCompile(`^[a-zA-Z0-9]+$`).MatchString(createRequest.Nickname) {
		return fmt.Errorf(common.ERR_NICKNAME_CHARSET_MISMATCH)
	}

	// Check the nick's length contraints.
	if len(createRequest.Nickname) > 12 || len(createRequest.Nickname) < 3 {
		return fmt.Errorf(common.ERR_NICKNAME_TOO_LONG_SHORT)
	}

	// Preprocess the e-mail address: set to lowercase.
	email := strings.ToLower(createRequest.Email)
	createRequest.Email = email

	// Validate the e-mail format.
	// https://stackoverflow.com/a/66624104
	if _, err := netmail.ParseAddress(createRequest.Email); err != nil {
		return fmt.Errorf(common.ERR_WRONG_EMAIL_FORMAT)
	}

	// Check if the e-mail address already used.
	users, err := s.userRepository.GetAll()
	if err != nil {
		return err
	}

	for key, dbUser := range *users {
		// Check if the nickname has been already used/taken.
		if key == createRequest.Nickname {
			return fmt.Errorf(common.ERR_USER_NICKNAME_TAKEN)
		}

		// E-mail address match found.
		if strings.ToLower(dbUser.Email) == createRequest.Email {
			return fmt.Errorf(common.ERR_EMAIL_ALREADY_USED)
		}
	}

	//
	//  Validation end
	//

	//
	//  Set user defaults, save the user struct to database and create a new system post
	//

	user := new(models.User)

	passHash := sha512.Sum512([]byte(createRequest.PassphrasePlain + os.Getenv("APP_PEPPER")))
	passHashHex := fmt.Sprintf("%x", passHash)

	// Transfer fields from the request to a new User object.
	user.Nickname = createRequest.Nickname
	user.PassphraseHex = passHashHex
	user.Email = createRequest.Email

	user.FlowList = make(models.UserGenericMap)
	user.FlowList[createRequest.Nickname] = true
	user.FlowList["system"] = true

	// Set the defaults and a timestamp.
	user.RegisteredTime = time.Now()
	user.LastActiveTime = time.Now()
	user.About = "newbie"

	// Set the default avatar.
	user.AvatarURL = func() string {
		if os.Getenv("APP_URL_MAIN") != "" {
			return "https://" + os.Getenv("APP_URL_MAIN") + "/web/apple-touch-icon.png"
		}

		return "https://www.littr.eu/web/apple-touch-icon.png"
	}()

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
	msg, err := s.mailService.ComposeMail(mailPayload)
	if err != nil || msg == nil {
		return fmt.Errorf(common.ERR_MAIL_COMPOSITION_FAIL)
	}

	// Send the activation mail to such user.
	if err = s.mailService.SendMail(msg); err != nil {
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

func (s *UserService) Subscribe(ctx context.Context, device *models.Device) error {
	// Fetch the callerID from the given context.
	callerID := common.GetCallerID(ctx)

	// Check whether the given device is blank.
	if reflect.DeepEqual(*device, (models.Device{})) {
		return fmt.Errorf(common.ERR_DEVICE_BLANK)
	}

	dbUser, err := s.userRepository.GetByID(callerID)
	if err != nil {
		return err
	}

	for _, dev := range dbUser.Devices {
		if dev.UUID == device.UUID {
			return fmt.Errorf("%s", common.ERR_DEVICE_SUBSCRIBED_ALREADY)
		}
	}

	dbUser.Devices = append(dbUser.Devices, *device)

	if err := s.userRepository.Save(dbUser); err != nil {
		return err
	}

	return nil
}

func (s *UserService) Unsubscribe(ctx context.Context, uuid string) error {
	// Fetch the callerID from the given context.
	callerID := common.GetCallerID(ctx)

	dbUser, err := s.userRepository.GetByID(callerID)
	if err != nil {
		return nil
	}

	var newDevices []models.Device

	for _, dev := range dbUser.Devices {
		if dev.UUID == uuid {
			continue
		}

		if reflect.DeepEqual(dev, (models.Device{})) {
			continue
		}

		if dev.UUID == "" {
			continue
		}

		newDevices = append(newDevices, dev)
	}

	dbUser.Devices = newDevices

	if err := s.userRepository.Save(dbUser); err != nil {
		return err
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

func (s *UserService) Update(ctx context.Context, userID, reqType string, userRequest interface{}) error {
	// Fetch the callerID from the given context.
	callerID := common.GetCallerID(ctx)

	switch reqType {
	case "lists":
		// Assert the type for the user update request.
		data, ok := userRequest.(*UserUpdateListsRequest)
		if !ok {
			return ErrUserRequestDecodingFailed
		}

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

			/*for key, value := range data.FlowList {
				if value && dbUser.FlowList[key] != data.FlowList[key] {
					return fmt.Errorf(common.ERR_USER_SHADED)
				}
			}*/
		}

		// Process the requestList request.
		if data.RequestList != nil {
			dbUser = processRequestList(data, dbUser, caller)

			for key, value := range data.RequestList {
				if value && dbUser.RequestList[key] != data.RequestList[key] {
					return fmt.Errorf(common.ERR_USER_SHADED)
				}
			}
		}

		// Process the shadeList request.
		if data.ShadeList != nil {
			dbUser = processShadeList(data, dbUser, caller)
		}

		if err := s.userRepository.Save(dbUser); err != nil {
			return err
		}

	case "options":
		// Assert the type for the user update request.
		data, ok := userRequest.(*UserUpdateOptionsRequest)
		if !ok {
			return ErrUserRequestDecodingFailed
		}

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
		if data.UIMode != dbUser.UIMode {
			dbUser.UIMode = !dbUser.UIMode
			dbUser.Options["uiMode"] = data.UIMode
		}
		if data.UITheme != dbUser.UITheme {
			dbUser.UITheme = data.UITheme
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
		// Assert the type for the user update request.
		data, ok := userRequest.(*UserUpdatePassphraseRequest)
		if !ok {
			return ErrUserRequestDecodingFailed
		}

		if callerID != userID {
			return fmt.Errorf(common.ERR_USER_PASSPHRASE_FOREIGN)
		}

		// Fetch requested user from repository.
		dbUser, err := s.userRepository.GetByID(userID)
		if err != nil {
			return fmt.Errorf(common.ERR_USER_NOT_FOUND)
		}

		// Check if both new or old passphrase hashes are blank/empty.
		if data.NewPassphrase == "" || data.CurrentPassphrase == "" {
			return fmt.Errorf(common.ERR_PASSPHRASE_REQ_INCOMPLETE)
		}

		pepper := os.Getenv("APP_PEPPER")

		passHashNew := sha512.Sum512([]byte(data.NewPassphrase + pepper))
		passHashCurrent := sha512.Sum512([]byte(data.CurrentPassphrase + pepper))

		// Check if the current passphrasë́'s hash is correct.
		if fmt.Sprintf("%x", passHashCurrent) != dbUser.PassphraseHex {
			return fmt.Errorf(common.ERR_PASSPHRASE_CURRENT_WRONG)
		}

		// Update user's passphrase.
		dbUser.PassphraseHex = fmt.Sprintf("%x", passHashNew)

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
	data, ok := userRequest.(*UserUploadAvatarRequest)
	if !ok {
		return nil, ErrUserRequestDecodingFailed
	}

	// Fetch the callerID/nickname type from the given context.
	callerID := common.GetCallerID(ctx)

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

	// Prepare the avatarURL to delete the previous avatar (if an uploaded image).
	prevAvatarURL := user.AvatarURL

	// Create a new regexp object to be used on the prevAvatarURL variable.
	regex, err := regexp.Compile("^(http|https)://")
	if err != nil {
		return nil, fmt.Errorf(common.ERR_AVATAR_URL_REGEXP_FAIL)
	}

	// Delete the saved avatar from the filesystem.
	if !regex.MatchString(prevAvatarURL) {
		fileName := strings.Replace(prevAvatarURL, "/web/", "/opt/", 1)

		// Do not fail on missing file: this prevents uploading a new avatar when
		// the current one is missing in the filesystem...
		_ = os.Remove(fileName)
	}

	user.AvatarURL = "/web/pix/thumb_" + *imageBaseURL

	// Update user's data.
	err = s.userRepository.Save(user)
	if err != nil {
		return nil, err
	}

	return imageBaseURL, nil
}

func (s *UserService) UpdateSubscriptionTags(ctx context.Context, uuid string, tags []string) error {
	callerID := common.GetCallerID(ctx)

	dbUser, err := s.userRepository.GetByID(callerID)
	if err != nil {
		return nil
	}

	var (
		devIdx int
		found  bool
	)

	for idx, dev := range dbUser.Devices {
		if dev.UUID == uuid {
			found = true
			devIdx = idx
			break
		}
	}

	if !found {
		return fmt.Errorf("%s", common.ERR_SUBSCRIPTION_NOT_FOUND)
	}

	for _, tag := range tags {
		if !slices.Contains(dbUser.Devices[devIdx].Tags, tag) {
			dbUser.Devices[devIdx].Tags = append(dbUser.Devices[devIdx].Tags, tag)
		} else {
			var newTags []string

			for _, t := range dbUser.Devices[devIdx].Tags {
				if t != tag {
					newTags = append(newTags, t)
				}
			}

			dbUser.Devices[devIdx].Tags = newTags
		}
	}

	if err := s.userRepository.Save(dbUser); err != nil {
		return err
	}

	return nil
}

func (s *UserService) ProcessPassphraseRequest(ctx context.Context, requestType string, userRequest interface{}) error {
	if requestType == "" {
		return fmt.Errorf(common.ERR_REQUEST_TYPE_BLANK)
	}

	var mailType string
	var user *models.User

	var randomPassphrase string
	var randomUUID string

	switch requestType {
	case "request":
		// Assert the type for the user update request.
		data, ok := userRequest.(*UserPassphraseRequest)
		if !ok {
			return ErrUserRequestDecodingFailed
		}

		if data.Email == "" {
			return fmt.Errorf(common.ERR_REQUEST_EMAIL_BLANK)
		}

		users, err := s.userRepository.GetAll()
		if err != nil {
			return err
		}

		var found bool
		var dbUser models.User

		for _, user := range *users {
			if strings.EqualFold(data.Email, user.Email) {
				found = true
				dbUser = user
				break
			}
		}

		if !found {
			return fmt.Errorf(common.ERR_NO_EMAIL_MATCH)
		}

		randomID := uuid.New().String()

		// Prepare the request data for the database.
		request := &models.Request{
			ID:        randomID,
			Nickname:  dbUser.Nickname,
			Email:     strings.ToLower(data.Email),
			CreatedAt: time.Now(),
			Type:      "reset_passphrase",
		}

		if err := s.requestRepository.Save(request); err != nil {
			return err
		}

		mailType = "reset_request"
		user = &dbUser

		randomUUID = randomID

	case "reset":
		// Assert the type for the user update request.
		data, ok := userRequest.(*UserPassphraseReset)
		if !ok {
			return ErrUserRequestDecodingFailed
		}

		if data.UUID == "" {
			return fmt.Errorf(common.ERR_REQUEST_UUID_BLANK)
		}

		request, err := s.requestRepository.GetByID(data.UUID)
		if err != nil {
			return err
		}

		// Check the request's validity.
		if request.CreatedAt.Before(time.Now().Add(-24 * time.Hour)) {
			// Delete the invalid request from the database.
			if err := s.requestRepository.Delete(data.UUID); err != nil {
				return fmt.Errorf(common.ERR_REQUEST_DELETE_FAIL)
			}

			return fmt.Errorf(common.ERR_REQUEST_UUID_EXPIRED)
		}

		// Preprocess the e-mail address = use the lowecased form.
		email := strings.ToLower(request.Email)
		request.Email = email

		dbUser, err := s.userRepository.GetByID(request.Nickname)
		if err != nil {
			return err
		}

		// Reset the passphrase = generete a new one (32 chars long).
		randomPassphrase = helpers.RandSeq(32)
		pepper := config.ServerSecret

		if pepper == "" {
			return fmt.Errorf(common.ERR_NO_SERVER_SECRET)
		}

		// Convert new passphrase into the hexadecimal format with pepper added.
		passHash := sha512.Sum512([]byte(randomPassphrase + pepper))
		dbUser.PassphraseHex = fmt.Sprintf("%x", passHash)

		if err := s.userRepository.Save(dbUser); err != nil {
			return err
		}

		if err := s.requestRepository.Delete(data.UUID); err != nil {
			return err
		}

		mailType = "reset_passphrase"
		user = dbUser

	default:
		return fmt.Errorf(common.ERR_REQUEST_TYPE_UNKNOWN)
	}

	// Prepare the mail options.
	mailPayload := mail.MessagePayload{
		Nickname:   user.Nickname,
		Email:      user.Email,
		Type:       mailType,
		UUID:       randomUUID,
		Passphrase: randomPassphrase,
	}

	// Compose a message to send.
	msg, err := s.mailService.ComposeMail(mailPayload)
	if err != nil || msg == nil {
		return fmt.Errorf(common.ERR_MAIL_COMPOSITION_FAIL)
	}

	// send the message
	if err := s.mailService.SendMail(msg); err != nil {
		return fmt.Errorf(common.ERR_MAIL_NOT_SENT)
	}

	return nil
}

func (s *UserService) Delete(ctx context.Context, userID string) error {
	// Fetch the caller's ID from the context.
	callerID := common.GetCallerID(ctx)

	// Fetch the user's ID from the context.
	if userID == "" {
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

	//
	//  Delete all posts, delete polls, delete tokens
	//

	polls, err := s.pollRepository.GetAll()
	if err != nil {
		return err
	}

	posts, err := s.postRepository.GetAll()
	if err != nil {
		return err
	}

	tokens, err := s.tokenRepository.GetAll()
	if err != nil {
		return err
	}

	// Spinoff a goroutine to process all the deletions async.
	go func(pollRepo models.PollRepositoryInterface, postRepo models.PostRepositoryInterface, tokenRepo models.TokenRepositoryInterface) {
		l := common.NewLogger(nil, "userDelete")

		for key, poll := range *polls {
			if poll.Author == userID {
				// Delete the poll.
				if err := pollRepo.Delete(key); err != nil {
					l.Msg("could not delete a poll: " + key).Status(http.StatusInternalServerError).Log()
					continue
				}
			}
		}

		for key, post := range *posts {
			if post.Nickname == userID {
				// Delete the post.
				if err := postRepo.Delete(key); err != nil {
					l.Msg("could not delete a post: " + key).Status(http.StatusInternalServerError).Log()
					continue
				}

				// Delete associated image and its thumbnail.
				if post.Figure != "" {
					err := os.Remove("/opt/pix/thumb_" + post.Figure)
					if err != nil {
						l.Msg(common.ERR_POST_DELETE_THUMB).Status(http.StatusInternalServerError).Error(err).Log()
						continue
					}

					err = os.Remove("/opt/pix/" + post.Figure)
					if err != nil {
						l.Msg(common.ERR_POST_DELETE_FULLIMG).Status(http.StatusInternalServerError).Error(err).Log()
						continue
					}
				}
			}
		}

		for key, token := range *tokens {
			if token.Nickname == userID {
				// Delete the token.
				if err := tokenRepo.Delete(key); err != nil {
					l.Msg("could not delete a token: " + key).Status(http.StatusInternalServerError).Log()
					continue
				}
			}
		}

		l.Msg("associated data linked to a just deleted user have been purged").Status(http.StatusOK).Log()
	}(s.pollRepository, s.postRepository, s.tokenRepository)

	return nil
}

func (s *UserService) FindAll(ctx context.Context, pageReq interface{}) (*map[string]models.User, error) {
	// Fetch the caller's ID from the context.
	callerID := common.GetCallerID(ctx)

	req, ok := pageReq.(*UserPagingRequest)
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
		PageNo:   req.PageNo,
		FlowList: nil,

		Users: pages.UserOptions{
			Plain:       true,
			RequestList: &caller.RequestList,
		},
	}

	allUsers, err := s.userRepository.GetAll()
	if err != nil {
		return nil, err
	}

	iface, err := s.pagingService.GetOne(ctx, opts, allUsers)
	if err != nil {
		return nil, err
	}

	ptrs, ok := iface.(pages.PagePointers)
	if !ok {
		return nil, fmt.Errorf("cannot assert type map of users")
	}

	// Patch the user's data for export.
	patchedUsers := common.FlushUserData(*&ptrs.Users, callerID)

	return patchedUsers, nil
}

func (s *UserService) FindByID(ctx context.Context, userID string) (*models.User, error) {
	if userID == "" {
		return nil, fmt.Errorf("userID argument blank")
	}

	// Fetch the caller's ID from the context.
	callerID := common.GetCallerID(ctx)

	// Request the user's data from repository..
	user, err := s.userRepository.GetByID(userID)
	if err != nil {
		return nil, err
	}

	// Include subscription devices if the userID is the caller's one.
	if userID != callerID {
		user.Devices = nil
	}

	// Patch the user's data for export.
	patchedUser := (*common.FlushUserData(&map[string]models.User{userID: *user}, userID))[userID]

	return &patchedUser, nil
}

func (s *UserService) FindPostsByID(ctx context.Context, userID string, pageOpts interface{}) (*map[string]models.Post, *map[string]models.User, error) {
	// Fetch the caller's ID from context.
	callerID := common.GetCallerID(ctx)

	caller, err := s.userRepository.GetByID(callerID)
	if err != nil {
		return nil, nil, err
	}

	req, ok := pageOpts.(*UserPagingRequest)
	if !ok {
		return nil, nil, fmt.Errorf(common.ERR_PAGENO_INCORRECT)
	}

	// Set the page options.
	opts := &pages.PageOptions{
		CallerID: callerID,
		PageNo:   req.PageNo,
		FlowList: nil,

		Flow: pages.FlowOptions{
			HideReplies:  req.HideReplies,
			Plain:        !req.HideReplies,
			UserFlow:     true,
			UserFlowNick: userID,
		},
	}

	allPosts, err := s.postRepository.GetAll()
	if err != nil {
		return nil, nil, err
	}

	iface, err := s.pagingService.GetOne(ctx, opts, []any{allPosts})
	if err != nil {
		return nil, nil, err
	}

	ptrs, ok := iface.(pages.PagePointers)
	if !ok {
		return nil, nil, fmt.Errorf(common.ERR_PAGE_EXPORT_NIL)
	}

	var users *map[string]models.User
	*users = make(map[string]models.User)
	(*users)[callerID] = *caller

	// Patch the user data export.
	users = common.FlushUserData(users, callerID)

	return ptrs.Posts, users, nil
}

//
//  Helpers
//

func processFlowList(data *UserUpdateListsRequest, user *models.User, caller *models.User, r models.UserRepositoryInterface) *models.User {
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
		if current, found := user.FlowList[key]; found {
			if current == value {
				continue
			}

			user.FlowList[key] = value
		}

		// Check if the caller is shaded by the counterpart.
		counterpart, err := r.GetByID(key)
		if err == nil {
			if counterpart.Private && key != caller.Nickname {
				// Cannot add this user to one's flow, as the following
				// has to be requested and allowed by the counterpart.
				user.FlowList[key] = false
				continue
			}

			if counterpart.ShadeList == nil {
				user.FlowList[key] = value
				continue
			}

			// Update the flowList record according to the counterpart's shade list state of the user.
			shaded, found := counterpart.ShadeList[caller.Nickname]
			if value && found && shaded {
				user.FlowList[key] = false
				continue
			}

			if !found || (found && !shaded) {
				user.FlowList[key] = value
			}
		}

		// Do not allow to unfollow oneself.
		if key == caller.Nickname && user.Nickname == caller.Nickname {
			user.FlowList[key] = true
		}

	}
	// Always allow to see system posts.
	user.FlowList["system"] = true

	return user
}

// Simple args legend: <data> coming from the <caller>'s side, <user> is to be updated as the primary counterpart.
// The <user> counterpart is barely equal to the <caller> part. Therefore the logic is reversed in comparison to the FlowList logic.
func processRequestList(data *UserUpdateListsRequest, user *models.User, caller *models.User) *models.User {
	if user.RequestList == nil {
		user.RequestList = make(map[string]bool)
	}

	// Loop over the RequestList records and change the user's values accordingly (enforce the proper requestList changing!).
	for key, value := range data.RequestList {
		// Do not request following on oneself.
		if user.Nickname == caller.Nickname {
			user.RequestList[key] = false
			continue
		}

		// Caller can only change the status on theirselves at the counterpart's.
		if key == caller.Nickname {
			// If the caller is not shaded (the list is empty/nil), proceed and assign the posted value.
			if len(user.ShadeList) == 0 {
				user.RequestList[key] = value
				continue
			}

			// Check the shade state for the <key> user.
			shaded, found := user.ShadeList[key]
			// The caller is shaded, so disallow any such request.
			if found && shaded {
				user.RequestList[key] = false
				continue
			}

			// User is not shaded => procced and assign the posted value.
			if !found || (found && !shaded) {
				user.RequestList[key] = value
			}
		}
	}

	return user
}

func processShadeList(data *UserUpdateListsRequest, user *models.User, caller *models.User) *models.User {
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
