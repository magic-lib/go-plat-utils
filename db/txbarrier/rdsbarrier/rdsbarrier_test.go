package rdsbarrier

import (
	"context"
	"strconv"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"

	"github.com/magic-lib/go-plat-utils/db/txbarrier"
)

const testScript = `
local v = redis.call('GET', KEYS[1])

if v == false or v + ARGV[1] < 0 then
	return {'', 'FAILURE'}
end

local ret = redis.call('INCRBY', KEYS[1], ARGV[1])
return {ret, 'SUCCESS'}
`

const testBranch = "deduction"

func TestRDSBarrierHook(t *testing.T) {
	t.Run("test: normal", func(t *testing.T) {
		s := miniredis.RunT(t)
		_ = s.Set("balance", "30") // init balance
		cli := redis.NewUniversalClient(&redis.UniversalOptions{Addrs: []string{s.Addr()}})
		cli.AddHook(NewHook(WithTimeout(3600)))

		// Try
		ctx := txbarrier.NewCtxWithBarrier(context.TODO(), &txbarrier.Barrier{
			XID:      "1",
			BranchID: testBranch,
			TransTyp: "tcc",
			Op:       txbarrier.Try,
		})
		ret, err := cli.Eval(ctx, testScript, []string{"balance"}, -1).Slice()
		require.Nil(t, err)
		require.Equal(t, []any{int64(29), "SUCCESS"}, ret)

		// Confirm
		ctx = txbarrier.NewCtxWithBarrier(context.TODO(), &txbarrier.Barrier{
			XID:      "1",
			BranchID: testBranch,
			TransTyp: "tcc",
			Op:       txbarrier.Confirm,
		})
		ret, err = cli.Eval(ctx, testScript, []string{"balance"}, 0).Slice()
		require.Nil(t, err)
		require.Equal(t, []any{int64(29), "SUCCESS"}, ret)
	})
	t.Run("test: cancel", func(t *testing.T) {
		s := miniredis.RunT(t)
		_ = s.Set("balance", "30") // init balance
		cli := redis.NewUniversalClient(&redis.UniversalOptions{Addrs: []string{s.Addr()}})
		cli.AddHook(NewHook(WithTimeout(3600)))

		// Try
		ctx := txbarrier.NewCtxWithBarrier(context.TODO(), &txbarrier.Barrier{
			XID:      "2",
			BranchID: testBranch,
			TransTyp: "tcc",
			Op:       txbarrier.Try,
		})
		ret, err := cli.Eval(ctx, testScript, []string{"balance"}, -1).Slice()
		require.Nil(t, err)
		require.Equal(t, []any{int64(29), "SUCCESS"}, ret)

		// Cancel
		ctx = txbarrier.NewCtxWithBarrier(context.TODO(), &txbarrier.Barrier{
			XID:      "2",
			BranchID: testBranch,
			TransTyp: "tcc",
			Op:       txbarrier.Cancel,
		})
		ret, err = cli.Eval(ctx, testScript, []string{"balance"}, 1).Slice()
		require.Nil(t, err)
		require.Equal(t, []any{int64(30), "SUCCESS"}, ret)
	})
	t.Run("test: idempotent", func(t *testing.T) {
		s := miniredis.RunT(t)
		_ = s.Set("balance", "30") // init balance
		cli := redis.NewUniversalClient(&redis.UniversalOptions{Addrs: []string{s.Addr()}})
		cli.AddHook(NewHook(WithTimeout(3600)))

		// Try
		ctx := txbarrier.NewCtxWithBarrier(context.TODO(), &txbarrier.Barrier{
			XID:      "3",
			BranchID: testBranch,
			TransTyp: "tcc",
			Op:       txbarrier.Try,
		})
		ret, err := cli.Eval(ctx, testScript, []string{"balance"}, -1).Slice()
		require.Nil(t, err)
		require.Equal(t, []any{int64(29), "SUCCESS"}, ret)

		// retry
		ctx = txbarrier.NewCtxWithBarrier(context.TODO(), &txbarrier.Barrier{
			XID:      "3",
			BranchID: testBranch,
			TransTyp: "tcc",
			Op:       txbarrier.Try,
		})
		_, err = cli.Eval(ctx, testScript, []string{"balance"}, -1).Slice()
		require.Equal(t, txbarrier.ErrDuplicationOrSuspension, err)
	})
	t.Run("test: empty compensation and hanging request", func(t *testing.T) {
		s := miniredis.RunT(t)
		_ = s.Set("balance", "30") // init balance
		cli := redis.NewUniversalClient(&redis.UniversalOptions{Addrs: []string{s.Addr()}})
		cli.AddHook(NewHook(WithTimeout(3600)))

		// Cancel: empty compensation
		ctx := txbarrier.NewCtxWithBarrier(context.TODO(), &txbarrier.Barrier{
			XID:      "4",
			BranchID: testBranch,
			TransTyp: "tcc",
			Op:       txbarrier.Cancel,
		})
		_, err := cli.Eval(ctx, testScript, []string{"balance"}, 1).Slice()
		require.Equal(t, txbarrier.ErrEmptyCompensation, err)

		// Try: hanging request
		ctx = txbarrier.NewCtxWithBarrier(context.TODO(), &txbarrier.Barrier{
			XID:      "4",
			BranchID: testBranch,
			TransTyp: "tcc",
			Op:       txbarrier.Try,
		})
		_, err = cli.Eval(ctx, testScript, []string{"balance"}, -1).Slice()
		require.Equal(t, txbarrier.ErrDuplicationOrSuspension, err)
	})
}

func TestRDSBarrierHookEvalsha(t *testing.T) {
	t.Run("test: evalsha normal", func(t *testing.T) {
		s := miniredis.RunT(t)
		_ = s.Set("balance", "30") // init balance
		cli := redis.NewUniversalClient(&redis.UniversalOptions{Addrs: []string{s.Addr()}})
		cli.AddHook(NewHook(WithTimeout(3600)))

		scriptSha, err := cli.ScriptLoad(context.TODO(), WrapScript(testScript)).Result()
		require.Nil(t, err)

		// Try
		ctx := txbarrier.NewCtxWithBarrier(context.TODO(), &txbarrier.Barrier{
			XID:      "1",
			BranchID: testBranch,
			TransTyp: "tcc",
			Op:       txbarrier.Try,
		})
		ret, err := cli.EvalSha(ctx, scriptSha, []string{"balance"}, -1).Slice()
		require.Nil(t, err)
		require.Equal(t, []any{int64(29), "SUCCESS"}, ret)

		// Confirm
		ctx = txbarrier.NewCtxWithBarrier(context.TODO(), &txbarrier.Barrier{
			XID:      "1",
			BranchID: testBranch,
			TransTyp: "tcc",
			Op:       txbarrier.Confirm,
		})
		ret, err = cli.EvalSha(ctx, scriptSha, []string{"balance"}, 0).Slice()
		require.Nil(t, err)
		require.Equal(t, []any{int64(29), "SUCCESS"}, ret)
	})
}

func TestHook_buildNewKeys(t *testing.T) {
	b := &txbarrier.Barrier{
		XID:      "1",
		BranchID: testBranch,
		TransTyp: "tcc",
		Op:       txbarrier.Cancel,
	}

	t.Run("test: simple barrier data key", func(t *testing.T) {
		h, _ := NewHook().(*Hook)
		ret, err := h.buildNewKeys(b, nil)
		require.Nil(t, err)
		require.Equal(t, []string{"1_deduction_cancel", "1_deduction_try"}, ret)
	})
	t.Run("test: empty key in cluster mode", func(t *testing.T) {
		h, _ := NewHook(WithClusterMode(true)).(*Hook)
		ret, err := h.buildNewKeys(b, nil)
		require.Contains(t, err.Error(), "must not be empty in cluster mode")
		require.Nil(t, ret)
	})
	t.Run("test: one key in cluster mode", func(t *testing.T) {
		h, _ := NewHook(WithClusterMode(true)).(*Hook)
		ret, err := h.buildNewKeys(b, []string{"testKey"})
		require.Nil(t, err)
		require.Equal(t, []string{"{testKey}_1_deduction_cancel", "{testKey}_1_deduction_try", "testKey"}, ret)

		ret, err = h.buildNewKeys(b, []string{"{tag}_testKey"})
		require.Nil(t, err)
		require.Equal(t, []string{"{tag}_1_deduction_cancel", "{tag}_1_deduction_try", "{tag}_testKey"}, ret)
	})
	t.Run("test: multi key in cluster mode", func(t *testing.T) {
		h, _ := NewHook(WithClusterMode(true)).(*Hook)
		ret, err := h.buildNewKeys(b, []string{"{tag}_testKey1", "{tag}_testKey2"})
		require.Nil(t, err)
		require.Equal(t, []string{"{tag}_1_deduction_cancel", "{tag}_1_deduction_try",
			"{tag}_testKey1", "{tag}_testKey2"}, ret)
	})
	t.Run("test: has no tag in multi key in cluster mode", func(t *testing.T) {
		h, _ := NewHook(WithClusterMode(true)).(*Hook)
		ret, err := h.buildNewKeys(b, []string{"testKey1", "testKey2"})
		require.Contains(t, err.Error(), "lack of hash tag in cluster mode")
		require.Nil(t, ret)
	})
}

func BenchmarkHook(b *testing.B) {
	s := miniredis.NewMiniRedis()
	if err := s.Start(); err != nil {
		b.Fatalf("start miniredis error: %v", err)
	}
	defer s.Close()

	_ = s.Set("balance", "0") // init balance
	cli := redis.NewUniversalClient(&redis.UniversalOptions{Addrs: []string{s.Addr()}})
	cli.AddHook(NewHook(WithTimeout(3600)))

	for i := 0; i < b.N; i++ {
		xid := strconv.Itoa(i)
		// Try
		ctx := txbarrier.NewCtxWithBarrier(context.TODO(), &txbarrier.Barrier{
			XID:      xid,
			BranchID: testBranch,
			TransTyp: "tcc",
			Op:       txbarrier.Try,
		})
		_, err := cli.Eval(ctx, testScript, []string{"balance"}, 1).Slice()
		if err != nil {
			b.Fatalf("try failed, xid: %v, error: %v", xid, err)
		}

		// Confirm
		ctx = txbarrier.NewCtxWithBarrier(context.TODO(), &txbarrier.Barrier{
			XID:      xid,
			BranchID: testBranch,
			TransTyp: "tcc",
			Op:       txbarrier.Confirm,
		})
		_, err = cli.Eval(ctx, testScript, []string{"balance"}, 0).Slice()
		if err != nil {
			b.Fatalf("confirm failed, xid: %v, error: %v", xid, err)
		}
	}
}

func BenchmarkWithoutHook(b *testing.B) {
	s := miniredis.NewMiniRedis()
	if err := s.Start(); err != nil {
		b.Fatalf("start miniredis error: %v", err)
	}
	defer s.Close()

	_ = s.Set("balance", "0") // init balance
	cli := redis.NewUniversalClient(&redis.UniversalOptions{Addrs: []string{s.Addr()}})

	for i := 0; i < b.N; i++ {
		xid := strconv.Itoa(i)
		// Try
		ctx := txbarrier.NewCtxWithBarrier(context.TODO(), &txbarrier.Barrier{
			XID:      xid,
			BranchID: testBranch,
			TransTyp: "tcc",
			Op:       txbarrier.Try,
		})
		_, err := cli.Eval(ctx, testScript, []string{"balance"}, 1).Slice()
		if err != nil {
			b.Fatalf("try failed, xid: %v, error: %v", xid, err)
		}

		// Confirm
		ctx = txbarrier.NewCtxWithBarrier(context.TODO(), &txbarrier.Barrier{
			XID:      xid,
			BranchID: testBranch,
			TransTyp: "tcc",
			Op:       txbarrier.Confirm,
		})
		_, err = cli.Eval(ctx, testScript, []string{"balance"}, 0).Slice()
		if err != nil {
			b.Fatalf("confirm failed, xid: %v, error: %v", xid, err)
		}
	}
}
