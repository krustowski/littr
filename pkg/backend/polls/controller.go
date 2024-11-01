package polls

import (
	"net/http"

	"go.vxn.dev/littr/pkg/backend/common"
	//"go.vxn.dev/littr/pkg/models"
	//chi "github.com/go-chi/chi/v5"
)

type PollController struct {
	pollService PollServiceInterface
}

func NewPollController(service PollServiceInterface) *PollController {
	return &PollController{
		pollService: service,
	}
}

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
	l := common.NewLogger(r, "polls")

	l.Msg("new poll created successfully").Status(http.StatusOK).Log().Payload(nil).Write(w)
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
	l := common.NewLogger(r, "polls")

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
	l := common.NewLogger(r, "polls")

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
// @Success      200  {object}   common.APIResponse{data=polls.getPolls.responseData}
// @Failure      400  {object}   common.APIResponse
// @Failure      500  {object}   common.APIResponse
// @Router       /polls [get]
func (c *PollController) GetAll(w http.ResponseWriter, r *http.Request) {
	l := common.NewLogger(r, "polls")

	l.Msg("listing all polls").Status(http.StatusOK).Log().Payload(nil).Write(w)
	return
}

// GetByID return just one specified poll.
//
// @Summary      Get single poll
// @Description  get single poll
// @Tags         polls
// @Produce      json
// @Param        pollID path string true "poll ID"
// @Success      200  {object}  common.APIResponse{data=polls.getSinglePoll.responseData}
// @Failure      400  {object}  common.APIResponse
// @Failure      404  {object}  common.APIResponse
// @Router       /polls/{pollID} [get]
func (c *PollController) GetByID(w http.ResponseWriter, r *http.Request) {
	l := common.NewLogger(r, "polls")

	l.Msg("returning the requested poll's data").Status(http.StatusOK).Log().Payload(nil).Write(w)
	return
}
