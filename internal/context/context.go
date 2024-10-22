package context

type CtxKey int

const (
	CtxKeyUser CtxKey = iota
	CtxKeyConnection
)
