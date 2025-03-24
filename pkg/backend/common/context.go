package common

import "context"

// Used in context as a custom type.
type UserNickname string

const (
	ContextUserKeyName UserNickname = "nickname"
	DefaultCallerID    string       = "system"
)

func GetCallerID(ctx context.Context) string {
	callerID, ok := ctx.Value(ContextUserKeyName).(string)
	if !ok || callerID == "" {
		return DefaultCallerID
	}

	return callerID
}
