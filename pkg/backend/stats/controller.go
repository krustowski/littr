package stats

import (
	"net/http"

	"go.vxn.dev/littr/pkg/backend/common"
	"go.vxn.dev/littr/pkg/models"
)

type StatController struct {
	statService models.StatServiceInterface
}

func NewStatController(statService models.StatServiceInterface) *StatController {
	if statService == nil {
		return nil
	}

	return &StatController{
		statService: statService,
	}
}

// GetAll acts like a handler for stats page.
//
//	@Summary		Get stats
//	@Description		This function call retrieves the system and users' statistics.
//	@Tags			stats
//	@Produce		json
//	@Success		200	{object}	common.APIResponse{data=stats.GetAll.responseData}	"Stats were calculated and are returned."
//	@Failure		400	{object}	common.APIResponse{data=models.Stub}			"Invalid called ID."
//	@Failure		500	{object}	common.APIResponse{data=models.Stub}			"A serious problem occurred while stats were being calculated."
//	@Router			/stats [get]
func (c *StatController) GetAll(w http.ResponseWriter, r *http.Request) {
	l := common.NewLogger(r, "stats")

	type responseData struct {
		FlowStats map[string]int64           `json:"flow_stats" example:"online:3,users:5"`
		UserStats map[string]models.UserStat `json:"user_stats"`
		Users     map[string]models.User     `json:"users"`
	}

	// Skip blank callerID.
	if l.CallerID() == "" {
		l.Msg(common.ERR_CALLER_BLANK).Status(http.StatusBadRequest).Log().Payload(nil).Write(w)
		return
	}

	flowStats, userStats, users, err := c.statService.Calculate(r.Context())
	if err != nil {
		l.Msg(err.Error()).Status(http.StatusInternalServerError).Error(err).Log().Payload(nil).Write(w)
		return
	}

	pl := &responseData{
		FlowStats: *flowStats,
		UserStats: *userStats,
		Users:     *common.FlushUserData(users, l.CallerID()),
	}

	l.Msg("ok, dumping user and system stats").Status(http.StatusOK).Log().Payload(pl).Write(w)
	return
}
