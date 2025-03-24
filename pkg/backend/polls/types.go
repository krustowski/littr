package polls

type PollCreateRequest struct {
	// Question is to describe the main purpose of such poll.
	Question string `json:"question" example:"which one is your favourite?"`

	// OptionOne is the answer numero uno.
	OptionOne string `json:"option_one" example:"apple"`

	// OptionTwo is the answer numero dos.
	OptionTwo string `json:"option_two" example:"banana"`

	// OptionThree is the answer numero tres.
	OptionThree string `json:"option_three" example:"cashew"`
}

type PollPagingRequest struct {
	PageNo     int
	PagingSize int
}

type PollUpdateRequest struct {
	// The poll's ID is specified using an URL param.
	// The purpose of this field is to transfer the ID from the controller to the service in a more smooth way.
	ID string `json:"-" example:"1234567890000" swaggerignore:"true"`

	// These fields are to be filled in the request body data.
	OptionOneCount   int64 `json:"option_one_count" example:"3"`
	OptionTwoCount   int64 `json:"option_two_count" example:"2"`
	OptionThreeCount int64 `json:"option_three_count" example:"6"`
}
