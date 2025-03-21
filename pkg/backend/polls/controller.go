package polls

import (
	"context"
	"net/http"
	"strconv"

	"go.vxn.dev/littr/pkg/backend/common"
	"go.vxn.dev/littr/pkg/models"

	chi "github.com/go-chi/chi/v5"
)

const (
	LOGGER_WORKER_NAME string = "pollController"
)

type PollController struct {
	pollService models.PollServiceInterface
}

func NewPollController(pollService models.PollServiceInterface) *PollController {
	if pollService == nil {
		return nil
	}

	return &PollController{
		pollService: pollService,
	}
}

// addNewPoll ensures a new polls is created and saved.
//
//	@Summary		Add new poll
//	@Description		This function call handles a new poll request to the poll service, where new poll creation is ensured.
//	@Tags			polls
//	@Accept			json
//	@Produce		json
//	@Param			request	body		polls.PollCreateRequest			true	"A new poll's simplified body."
//	@Success		201		{object}	common.APIResponse{data=models.Stub}	"A new poll has been created successfully."
//	@Failure		400		{object}	common.APIResponse{data=models.Stub}	"Invalid input data."
//	@Failure		401		{object}	common.APIResponse{data=models.Stub}	"User unauthorized."
//	@Failure		429		{object}	common.APIResponse{data=models.Stub}	"Too many requests, try again later."
//	@Failure		500		{object}	common.APIResponse{data=models.Stub}	"A serious problem occurred while processing the create request."
//	@Router			/polls [post]
func (c *PollController) Create(w http.ResponseWriter, r *http.Request) {
	l := common.NewLogger(r, LOGGER_WORKER_NAME)

	// Skip the blank caller's ID.
	if l.CallerID() == "" {
		l.Msg(common.ERR_CALLER_BLANK).Status(http.StatusBadRequest).Log().Payload(nil).Write(w)
		return
	}

	var dtoIn PollCreateRequest

	// Decode the received data.
	if err := common.UnmarshalRequestData(r, &dtoIn); err != nil {
		l.Msg(common.ERR_INPUT_DATA_FAIL).Status(http.StatusBadRequest).Error(err).Log().Payload(nil).Write(w)
		return
	}

	// Create the poll at pollService.
	if err := c.pollService.Create(r.Context(), &dtoIn); err != nil {
		l.Msg(err.Error()).Status(common.DecideStatusFromError(err)).Log().Payload(nil).Write(w)
		return
	}

	l.Msg("new poll created successfully").Status(http.StatusCreated).Log().Payload(nil).Write(w)
}

// Update updates a given poll.
//
//	@Summary		Update a poll
//	@Description		This function call updates the poll specified using the `pollID` parameter. The fields to be updated are the counts of the poll's options. Only a single incrementation related to the current state stored in the database is allowed to be processed.
//	@Tags			polls
//	@Accept			json
//	@Produce		json
//	@Param			updatedPoll	body		polls.PollUpdateRequest	true		"A poll's body to update."
//	@Param			pollID		path		string		true			"A poll's unique ID."
//	@Success		200		{object}	common.APIResponse{data=models.Stub}	"The poll has been updated successfully."
//	@Failure		400		{object}	common.APIResponse{data=models.Stub}	"Invalid input data."
//	@Failure		401		{object}	common.APIResponse{data=models.Stub}	"User unauthorized."
//	@Failure		429		{object}	common.APIResponse{data=models.Stub}	"Too many requests, try again later."
//	@Failure		500		{object}	common.APIResponse{data=models.Stub}	"A serious internal problem occurred while the request was being processed."
//	@Router			/polls/{pollID} [patch]
func (c *PollController) Update(w http.ResponseWriter, r *http.Request) {
	l := common.NewLogger(r, LOGGER_WORKER_NAME)

	// Skip the blank caller's ID.
	if l.CallerID() == "" {
		l.Msg(common.ERR_CALLER_BLANK).Status(http.StatusBadRequest).Log().Payload(nil).Write(w)
		return
	}

	// Take the param from the URI path.
	pollID := chi.URLParam(r, "pollID")
	if pollID == "" {
		l.Msg(common.ERR_POLLID_BLANK).Status(http.StatusBadRequest).Log().Payload(nil).Write(w)
		return
	}

	var dtoIn PollUpdateRequest

	// Decode the received data.
	if err := common.UnmarshalRequestData(r, &dtoIn); err != nil {
		l.Msg(common.ERR_INPUT_DATA_FAIL).Status(http.StatusBadRequest).Log().Payload(nil).Write(w)
		return
	}

	// Thw two ways on how to pass the pollID value to the service.
	ctx := context.WithValue(r.Context(), "pollID", pollID)
	dtoIn.ID = pollID

	// Dispatch the update request to the pollService.
	if err := c.pollService.Update(ctx, &dtoIn); err != nil {
		l.Msg(err.Error()).Status(common.DecideStatusFromError(err)).Log().Payload(nil).Write(w)
		return
	}

	l.Msg("ok, poll has been updated successfully").Status(http.StatusOK).Log().Payload(nil).Write(w)
}

// Delete removes a poll.
//
//	@Summary		Delete a poll by ID
//	@Description		This function call takes in a `pollID` parameter, which is used to identify a poll to be purged.
//	@Tags			polls
//	@Produce		json
//	@Param			pollID	path		string		true			"A poll's ID to be deleted."
//	@Success		200	{object}	common.APIResponse{data=models.Stub}	"The poll has been deleted."
//	@Failure		400	{object}	common.APIResponse{data=models.Stub}	"Invalid input data."
//	@Failure		401	{object}	common.APIResponse{data=models.Stub}	"User unauthorized."
//	@Failure		403	{object}	common.APIResponse{data=models.Stub}	"User unauthorized. May occur when one tries to delete a foreign poll (the poll's author differs)."
//	@Failure		404	{object}	common.APIResponse{data=models.Stub}	"Poll not found in the database."
//	@Failure		429	{object}	common.APIResponse{data=models.Stub}	"Too many requests, try again later."
//	@Failure		500	{object}	common.APIResponse{data=models.Stub}	"A serious internal problem occurred while the request was being processed."
//	@Router			/polls/{pollID} [delete]
func (c *PollController) Delete(w http.ResponseWriter, r *http.Request) {
	l := common.NewLogger(r, LOGGER_WORKER_NAME)

	// Skip the blank caller's ID.
	if l.CallerID() == "" {
		l.Msg(common.ERR_CALLER_BLANK).Status(http.StatusBadRequest).Log().Payload(nil).Write(w)
		return
	}

	// Take the param from the URI path.
	pollID := chi.URLParam(r, "pollID")
	if pollID == "" {
		l.Msg(common.ERR_POLLID_BLANK).Status(http.StatusBadRequest).Log().Payload(nil).Write(w)
		return
	}

	// Dispatch the delete request to the pollService.
	if err := c.pollService.Delete(r.Context(), pollID); err != nil {
		l.Msg(err.Error()).Status(common.DecideStatusFromError(err)).Error(err).Log().Payload(nil).Write(w)
		return
	}

	// Log the message and write the HTTP response.
	l.Msg("ok, poll has been deleted successfully").Status(http.StatusOK).Log().Payload(nil).Write(w)
}

// GellAll gets a list of polls
//
//	@Summary		Get a list of polls
//	@Description		This function call retrieves a single page of polls according to the optional `X-Page-No` header (default is 0).
//	@Tags			polls
//	@Produce		json
//	@Param			X-Page-No	header		integer		false					"A page number (default is 0)."
//	@Success		200		{object}	common.APIResponse{data=polls.GetAll.responseData}	"The requested page of polls is returned."
//	@Failure		400		{object}	common.APIResponse{data=models.Stub}			"Invalid input data."
//	@Failure		401		{object}	common.APIResponse{data=models.Stub}			"User unauthorized."
//	@Failure		429		{object}	common.APIResponse{data=models.Stub}			"Too many requests, try again later."
//	@Failure		500		{object}	common.APIResponse{data=models.Stub}			"A serious internal problem occurred while the request was being processed."
//	@Router			/polls [get]
func (c *PollController) GetAll(w http.ResponseWriter, r *http.Request) {
	l := common.NewLogger(r, LOGGER_WORKER_NAME)

	// Skip the blank caller's ID.
	if l.CallerID() == "" {
		l.Msg(common.ERR_CALLER_BLANK).Status(http.StatusBadRequest).Log().Payload(nil).Write(w)
		return
	}

	// Fetch the pageNo from headers.
	pageNoString := r.Header.Get(common.HDR_PAGE_NO)
	pageNo, err := strconv.Atoi(pageNoString)
	if err != nil {
		//l.Msg(common.ERR_PAGENO_INCORRECT).Status(http.StatusBadRequest).Error(err).Log().Payload(nil).Write(w)
		//return
		pageNo = 0
	}

	type responseData struct {
		Polls *map[string]models.Poll `json:"polls,omitempty"`
		User  *models.User            `json:"user,omitempty"`
	}

	// Compose the DTO-out from pollService.
	polls, user, err := c.pollService.FindAll(context.WithValue(r.Context(), "pageNo", pageNo))
	if err != nil {
		l.Msg(err.Error()).Status(common.DecideStatusFromError(err)).Log().Payload(nil).Write(w)
		return
	}

	dtoOut := &responseData{Polls: polls, User: user}

	// Log the message and write the HTTP response.
	l.Msg("ok, listing all polls").Status(http.StatusOK).Log().Payload(dtoOut).Write(w)
}

// GetByID return just one specified poll.
//
//	@Summary		Get single poll
//	@Description		This function call retrieves a single requested poll's data. Such poll's ID is to be provided as the URL parameter.
//	@Tags			polls
//	@Produce		json
//	@Param			pollID	path	string	true							"A poll's ID to retrieve."
//	@Success		200	{object}	common.APIResponse{data=polls.GetByID.responseData}	"The requested poll's data returned successfully."
//	@Failure		400	{object}	common.APIResponse{data=models.Stub}			"Invalid input data."
//	@Failure		401	{object}	common.APIResponse{data=models.Stub}			"User unauthorized."
//	@Failure		404	{object}	common.APIResponse{data=models.Stub}			"Poll not found in the database."
//	@Failure		429	{object}	common.APIResponse{data=models.Stub}			"Too many requests, try again later."
//	@Failure		500	{object}	common.APIResponse{data=models.Stub}			"A serious internal problem occurred while the request was being processed."
//	@Router			/polls/{pollID} [get]
func (c *PollController) GetByID(w http.ResponseWriter, r *http.Request) {
	l := common.NewLogger(r, LOGGER_WORKER_NAME)

	// Skip the blank caller's ID.
	if l.CallerID() == "" {
		l.Msg(common.ERR_CALLER_BLANK).Status(http.StatusBadRequest).Log().Payload(nil).Write(w)
		return
	}

	// Take the param from the URI path.
	pollID := chi.URLParam(r, "pollID")
	if pollID == "" {
		l.Msg(common.ERR_POLLID_BLANK).Status(http.StatusBadRequest).Log().Payload(nil).Write(w)
		return
	}

	type responseData struct {
		Poll *models.Poll `json:"poll"`
		User *models.User `json:"user"`
	}

	// Compose the DTO-out from pollService.
	poll, user, err := c.pollService.FindByID(r.Context(), pollID)
	if err != nil {
		l.Msg(err.Error()).Status(common.DecideStatusFromError(err)).Log().Payload(nil).Write(w)
		return
	}

	dtoOut := &responseData{Poll: poll, User: user}

	// Log the message and write the HTTP response.
	l.Msg("ok, returning the requested poll's data").Status(http.StatusOK).Log().Payload(dtoOut).Write(w)
}
