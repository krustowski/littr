package config

var (
	// NicknameLengthMax is the maximum nickname length to be allowed to register.
	NicknameLengthMax int = 12

	// MaxPostLength is the maximal length of the fully shown post in the flow. Posts longer than that are to be shorten/hidden.
	MaxPostLength     int = 500
)
