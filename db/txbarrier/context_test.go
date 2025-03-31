package txbarrier

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBarrier_Valid(t *testing.T) {
	t.Run("invalid", func(t *testing.T) {
		var b *Barrier
		require.False(t, b.Valid())

		b = &Barrier{}
		require.False(t, b.Valid())
	})
	t.Run("valid", func(t *testing.T) {
		b := &Barrier{
			XID:      "1",
			BranchID: "2",
			TransTyp: "tcc",
			Op:       Try,
		}
		require.True(t, b.Valid())
	})
}

func TestSetAndGetBarrier(t *testing.T) {
	t.Run("set and get empty barrier", func(t *testing.T) {
		ctx := NewCtxWithBarrier(context.TODO(), nil)
		require.Nil(t, BarrierFromCtx(ctx))
	})
	t.Run("set and get barrier", func(t *testing.T) {
		ctx := NewCtxWithBarrier(context.TODO(), &Barrier{})
		require.NotNil(t, BarrierFromCtx(ctx))
	})

}
