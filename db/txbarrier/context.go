package txbarrier

import "context"

// Barrier defines the info which is needed for txbarrier.
type Barrier struct {
	XID      string
	BranchID string
	TransTyp string
	Op       Operation
}

var zerosBarrier = Barrier{}

// Valid returns whether the barrier info is valid.
func (b *Barrier) Valid() bool {
	return b != nil && *b != zerosBarrier
}

type barrierKey struct{}

// NewCtxWithBarrier creates a new context with Barrier.
func NewCtxWithBarrier(ctx context.Context, b *Barrier) context.Context {
	if b == nil {
		return ctx
	}
	return context.WithValue(ctx, barrierKey{}, b)
}

// BarrierFromCtx returns Barrier from the given context. It returns nil if
// ctx without Barrier.
func BarrierFromCtx(ctx context.Context) *Barrier {
	b, _ := ctx.Value(barrierKey{}).(*Barrier)
	return b
}
