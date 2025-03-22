package config

const (
	// NicknameLengthMax is the maximum nickname length to be allowed to register.
	MaxNicknameLength int = 12

	// MaxPostLength is the maximal length of the fully shown post in the flow. Posts longer than that are to be shorten/hidden.
	MaxPostLength int = 500

	// Max retry count for the minimal SSE client.
	MaxSseRetryCount int = 50
)
