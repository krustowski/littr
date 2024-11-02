package polls

import (
	"context"
	"net/http"
	"strconv"

	"go.vxn.dev/littr/pkg/backend/common"
	"go.vxn.dev/littr/pkg/models"

	chi "github.com/go-chi/chi/v5"
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

// Special vars to finger at the function reference of the PollController struct.
var createPollController = (&PollController{}).Create
var updatePollController = (&PollController{}).Update
var deletePollController = (&PollController{}).Delete
var getOnePollController = (&PollController{}).GetByID
var getAllPollController = (&PollController{}).GetAll

// addNewPoll ensures a new polls is created and saved.
//
// @Summary      Add new poll
// @Description  add new poll
// @Tags         polls
// @Accept       json
// @Produce      json
// @Param    	 request body models.Poll true "new poll's body"
// @Success      201  {object}  common.APIResponse "success"
// @Failure      400  {object}  common.APIResponse "bad/malformed input data, invalid cookies"
// @Failure      500  {object}  common.APIResponse "the poll saving process failed"
// @Router       /polls [post]
func (c *PollController) Create(w http.ResponseWriter, r *http.Request) {
	l := common.NewLogger(r, "pollController")

	// Skip the blank caller's ID.
	if l.CallerID() == "" {
		l.Msg(common.ERR_CALLER_BLANK).Status(http.StatusBadRequest).Log().Payload(nil).Write(w)
		return
	}

	var DTOIn models.Poll

	// Decode the received data.
	if err := common.UnmarshalRequestData(r, &DTOIn); err != nil {
		l.Msg(common.ERR_INPUT_DATA_FAIL).Status(http.StatusBadRequest).Error(err).Log()
		l.Msg(common.ERR_INPUT_DATA_FAIL).Status(http.StatusBadRequest).Payload(nil).Write(w)
		return
	}

	// Create the poll at pollService.
	if err := c.pollService.Create(r.Context(), &DTOIn); err != nil {
		l.Msg("could not create a new poll").Status(decideStatusFromError(err)).Error(err).Log()
		l.Msg("could not create a new poll").Status(decideStatusFromError(err)).Payload(nil).Write(w)
		return
	}

	l.Msg("new poll created successfully").Status(http.StatusCreated).Log().Payload(nil).Write(w)
	return
}

// Update updates a given poll.
//
// @Summary      Update a poll
// @Description  update a poll
// @Tags         polls
// @Accept       json
// @Produce      json
// @Param    	 updatedPoll body models.Poll true "update poll's body"
// @Param        pollID path string true "poll's ID"
// @Success      200  {object}  common.APIResponse
// @Failure      400  {object}  common.APIResponse
// @Failure      500  {object}  common.APIResponse
// @Router       /polls/{pollID} [put]
func (c *PollController) Update(w http.ResponseWriter, r *http.Request) {
	l := common.NewLogger(r, "pollController")

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

	var DTOIn *models.Poll

	// Decode the received data.
	if err := common.UnmarshalRequestData(r, DTOIn); err != nil {
		l.Msg(common.ERR_INPUT_DATA_FAIL).Status(http.StatusBadRequest).Error(err).Log()
		l.Msg(common.ERR_INPUT_DATA_FAIL).Status(http.StatusBadRequest).Payload(nil).Write(w)
		return
	}

	ctx := context.WithValue(r.Context(), "pollID", pollID)

	// Dispatch the update request to the pollService.
	if err := c.pollService.Update(ctx, DTOIn); err != nil {
		l.Msg("could not update the poll:").Status(decideStatusFromError(err)).Error(err).Log()
		l.Msg("could not update the poll:").Status(decideStatusFromError(err)).Payload(nil).Write(w)
		return
	}

	l.Msg("poll has been updated successfully").Status(http.StatusOK).Log().Payload(nil).Write(w)
	return
}

// Delete removes a poll.
//
// @Summary      Delete a poll by ID
// @Description  delete a poll by ID
// @Tags         polls
// @Produce      json
// @Param        pollID path string true "poll's ID to delete"
// @Success      200  {object}  common.APIResponse
// @Failure      400  {object}  common.APIResponse
// @Failure      403  {object}  common.APIResponse
// @Failure      500  {object}  common.APIResponse
// @Router       /polls/{pollID} [delete]
func (c *PollController) Delete(w http.ResponseWriter, r *http.Request) {
	l := common.NewLogger(r, "pollController")

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
		l.Msg("could not delete the poll").Status(decideStatusFromError(err)).Error(err).Log()
		l.Msg("could not delete the poll").Status(decideStatusFromError(err)).Payload(nil).Write(w)
		return
	}

	// Log the message and write the HTTP response.
	l.Msg("poll has been deleted successfully").Status(http.StatusOK).Log().Payload(nil).Write(w)
	return
}

// GellAll gets a list of polls
//
// @Summary      Get a list of polls
// @Description  get a list of polls
// @Tags         polls
// @Produce      json
// @Param    	 X-Page-No header string true "page number"
// @Success      200  {object}   common.APIResponse{data=polls.GetAll.responseData}
// @Failure      400  {object}   common.APIResponse
// @Failure      500  {object}   common.APIResponse
// @Router       /polls [get]
func (c *PollController) GetAll(w http.ResponseWriter, r *http.Request) {
	l := common.NewLogger(r, "pollController")

	// Skip the blank caller's ID.
	if l.CallerID() == "" {
		l.Msg(common.ERR_CALLER_BLANK).Status(http.StatusBadRequest).Log().Payload(nil).Write(w)
		return
	}

	// Fetch the pageNo from headers.
	pageNoString := r.Header.Get(common.HDR_PAGE_NO)
	pageNo, err := strconv.Atoi(pageNoString)
	if err != nil {
		l.Msg(common.ERR_PAGENO_INCORRECT).Status(http.StatusBadRequest).Error(err).Log()
		l.Msg(common.ERR_PAGENO_INCORRECT).Status(http.StatusBadRequest).Payload(nil).Write(w)
		return
	}

	type responseData struct {
		Polls *map[string]models.Poll `json:"polls,omitempty"`
		User  *models.User            `json:"user,omitempty"`
	}

	// Compose the DTO-out from pollService.
	polls, user, err := c.pollService.FindAll(context.WithValue(r.Context(), "pageNo", pageNo))
	if err != nil {
		l.Msg("could not fetch all polls").Status(decideStatusFromError(err)).Error(err).Log()
		l.Msg("could not fetch all polls").Status(decideStatusFromError(err)).Payload(nil).Write(w)
		return
	}

	DTOOut := &responseData{Polls: polls, User: user}

	// Log the message and write the HTTP response.
	l.Msg("listing all polls").Status(http.StatusOK).Log().Payload(DTOOut).Write(w)
	return
}

// GetByID return just one specified poll.
//
// @Summary      Get single poll
// @Description  get single poll
// @Tags         polls
// @Produce      json
// @Param        pollID path string true "poll ID"
// @Success      200  {object}  common.APIResponse{data=polls.GetByID.responseData}
// @Failure      400  {object}  common.APIResponse
// @Failure      404  {object}  common.APIResponse
// @Router       /polls/{pollID} [get]
func (c *PollController) GetByID(w http.ResponseWriter, r *http.Request) {
	l := common.NewLogger(r, "pollController")

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
		l.Msg("could not fetch requested poll").Status(decideStatusFromError(err)).Error(err).Log()
		l.Msg("could not fetch requested poll").Status(decideStatusFromError(err)).Payload(nil).Write(w)
		return
	}

	DTOOut := &responseData{Poll: poll, User: user}

	// Log the message and write the HTTP response.
	l.Msg("returning the requested poll's data").Status(http.StatusOK).Log().Payload(DTOOut).Write(w)
	return
}

//
//  Helpers
//

var decideStatusFromError = func(err error) int {
	if err == nil {
		return http.StatusOK
	}

	if err.Error() == common.ERR_POLL_NOT_FOUND {
		return http.StatusNotFound
	}

	return http.StatusInternalServerError
}
