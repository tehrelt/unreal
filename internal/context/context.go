package context

type CtxKey string

const (
	CtxKeyUser       CtxKey = "user"
	CtxKeyConnection CtxKey = "connection"
)
