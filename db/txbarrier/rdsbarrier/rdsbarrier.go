// Package rdsbarrier is designed to solve the timing problem of accessing
// RM(Resource Manager) based on redis in distributed transactions.
package rdsbarrier

import (
	"context"
	"fmt"
	"net"
	"strings"

	"github.com/redis/go-redis/v9"

	"github.com/magic-lib/go-plat-utils/db/txbarrier"
)

var cancelMapFirstPhase = map[txbarrier.Operation]txbarrier.Operation{
	txbarrier.Cancel: txbarrier.Try,
}

// key1: XID_BranchID_Op is the key inserted for the current operation.
//
// key2: XID_BranchID_Op is the key corresponding to the try operation that needs to be inserted due to an
// empty compensation, to prevent the try operation from reoccurring after an empty compensation.
//
// ARGV1: If the current operation is a cancel operation, ARGV1 is not empty.
//
// ARGV2: The duration for which the barrier data should be retained.
const barrierScript = `local function biz_logic(KEYS, ARGV)
%s
end

local e1 = redis.call('GET', KEYS[1])
if e1 ~= false then
	-- duplicated or hanging request
	return '%s'
end

if tonumber(ARGV[2]) > 0 then
	redis.call('SET', KEYS[1], 'op', 'EX', ARGV[2])
else
	redis.call('SET', KEYS[1], 'op')
end

-- Cancel operation, needs to check whether it's a empty compensation
if ARGV[1] ~= '' then
	local e2 = redis.call('GET', KEYS[2])
	if e2 == false then
		-- empty compensation
		if tonumber(ARGV[2]) > 0 then
			redis.call('SET', KEYS[2], 'rollback', 'EX', ARGV[2])
		else
			redis.call('SET', KEYS[2], 'rollback')
		end
		return '%s'
	end
end

local ret = biz_logic({unpack(KEYS, 3)}, {unpack(ARGV, 3)})
if ret[2] ~= 'SUCCESS' then
	-- biz logic failure, rollback barrier data.
    -- If the biz_logic operation involves business data, 
    -- it is the responsibility of the business to handle the rollback.
	redis.call('DEL', KEYS[1])
end

return ret
`

const (
	barrierResultDuplicate         = "DUPLICATE"
	barrierResultEmptyCompensation = "EMPTY_COMPENSATION"
)

// defaultBarrierDataTimeout is the default duration (seconds) for barrier data should be retained.
// The rdsbarrier will generate barrier data which is based on xid in redis to records the state of
// distributed transaction branch requests. By using barrier data, rdsbarrier can correctly handle
// duplicated, hanging and empty compensation requests.
//
// It's important to note that setting a too long timeout for barrier data can result in long-term
// occupation of redis memory, while a too short timeout may cause the barrier logic to fail, leading
// to incorrect protection of transaction branch requests.
//
// It is recommended for businesses to set a reasonable timeout for barrier data based on their needs,
// generally slightly longer than the overall timeout of the distributed transaction.
//
// Using WithTimeout to set timeout value for barrier data.
const defaultBarrierDataTimeout = 3600 * 24

// WrapScript wraps business script by barrier logic.
func WrapScript(script string) string {
	return fmt.Sprintf(barrierScript, script, barrierResultDuplicate, barrierResultEmptyCompensation)
}

// GetLogKey returns the key of barrier log in redis. The hashTag must be empty except
// when using redis cluster.
func GetLogKey(xid, branchID string, op txbarrier.Operation, hashTag string) string {
	if hashTag == "" {
		return fmt.Sprintf("%s_%s_%s", xid, branchID, op)
	}

	return fmt.Sprintf("{%s}_%s_%s_%s", hashTag, xid, branchID, op)
}

// Hook is a redis.Hook for assisting users to solve the timing problem
// of accessing RM(Resource Manager) in distributed transactions such as
// request idempotence, hanging requests, and empty compensation.
type Hook struct {
	timeout       int
	enableCluster bool
}

// NewHook creates a Hook for redis.
func NewHook(opts ...Option) redis.Hook {
	hook := &Hook{timeout: defaultBarrierDataTimeout}
	for _, o := range opts {
		o(hook)
	}

	return hook
}

// DialHook implements the redis.Hook. It does nothing currently.
func (*Hook) DialHook(next redis.DialHook) redis.DialHook {
	return func(ctx context.Context, network, addr string) (net.Conn, error) {
		return next(ctx, network, addr)
	}
}

// ProcessPipelineHook implements the redis.Hook. It does nothing currently.
func (*Hook) ProcessPipelineHook(next redis.ProcessPipelineHook) redis.ProcessPipelineHook {
	return func(ctx context.Context, cmd []redis.Cmder) error {
		return next(ctx, cmd)
	}
}

// ProcessHook implements the redis.Hook. In this method, the user's redis commands are
// intercepted, and barrier logic and data are inserted.
func (h *Hook) ProcessHook(next redis.ProcessHook) redis.ProcessHook {
	return func(ctx context.Context, cmd redis.Cmder) error {
		b := txbarrier.BarrierFromCtx(ctx)
		if cmd.Name() != "eval" && cmd.Name() != "evalsha" || !b.Valid() {
			// Unsupported redis commands, do nothing.
			return next(ctx, cmd)
		}

		// Extracts and parses the source eval script and its params.
		// According to github.com/redis/go-redis/v9 , we know cmd must be *redis.Cmd type.
		sc, _ := cmd.(*redis.Cmd)
		name := sc.Name()
		sPayload, sKeys, sArgs, err := h.extractEvalCmd(sc)
		if err != nil {
			return err
		}

		nPayload := sPayload
		if cmd.Name() == "eval" {
			// wraps business script by barrier.
			nPayload = WrapScript(sPayload)
		}

		// Builds new keys.
		nKeys, err := h.buildNewKeys(b, sKeys)
		if err != nil {
			return err
		}

		// Builds new args.
		firstPhaseOp := string(cancelMapFirstPhase[b.Op])
		nArgs := append([]any{firstPhaseOp, h.timeout}, sArgs...)

		// Builds new redis.Cmd and assigns it to the old one.
		nc := h.buildEvalCmd(ctx, name, nPayload, nKeys, nArgs)
		*sc = *nc

		if err = next(ctx, sc); err != nil {
			return err
		}
		return h.parseBarrierResult(sc)
	}
}

// extractEvalCmd is a reverse method of buildEvalCmd used to parse the argument of the "eval" command.
func (h *Hook) extractEvalCmd(cmd *redis.Cmd) (payload string, keys []string, args []any, err error) {
	payload, ok := cmd.Args()[1].(string)
	if !ok {
		err = fmt.Errorf("rdsbarrier: cmd.Args()[1] is %v, type %T, expected payload as string", cmd.Args()[1], cmd.Args()[1])
		return
	}

	keysLen, ok := cmd.Args()[2].(int)
	if !ok {
		err = fmt.Errorf("rdsbarrier: cmd.Args()[2] is %v, type %T, expected keysLen as int", cmd.Args()[2], cmd.Args()[2])
		return
	}

	keys = make([]string, 0, keysLen)
	for i := 0; i < keysLen; i++ {
		keys = append(keys, cmd.Args()[3+i].(string))
	}

	args = make([]any, 0, len(cmd.Args()[3+keysLen:]))
	args = append(args, cmd.Args()[3+keysLen:]...)
	return
}

func (h *Hook) buildNewKeys(b *txbarrier.Barrier, oldKeys []string) ([]string, error) {
	// if enable cluster mod, adds hash tag for keys.
	var tag string
	if h.enableCluster {
		if len(oldKeys) == 0 {
			return nil, fmt.Errorf("rdsbarrier: the key(s) of eval must not be empty in cluster mode")
		}

		l := strings.IndexByte(oldKeys[0], '{')
		r := strings.IndexByte(oldKeys[0], '}')
		if l == -1 || l >= r {
			if len(oldKeys) != 1 {
				return nil, fmt.Errorf("rdsbarrier: lack of hash tag in cluster mode")
			}

			tag = oldKeys[0]
		} else {
			tag = oldKeys[0][l+1 : r]
		}
	}

	newKeys := make([]string, 2, 2+len(oldKeys))

	// key1: XID_BranchID_Op
	newKeys[0] = GetLogKey(b.XID, b.BranchID, b.Op, tag)
	// key2: XID_BranchID_Op
	newKeys[1] = GetLogKey(b.XID, b.BranchID, cancelMapFirstPhase[b.Op], tag)
	newKeys = append(newKeys, oldKeys...)

	return newKeys, nil
}

// buildEvalCmd copies from redis.v9 package.
func (h *Hook) buildEvalCmd(ctx context.Context, name, payload string, keys []string, args []any) *redis.Cmd {
	cmdArgs := make([]any, 3+len(keys), 3+len(keys)+len(args))
	cmdArgs[0] = name
	cmdArgs[1] = payload
	cmdArgs[2] = len(keys)
	for i, key := range keys {
		cmdArgs[3+i] = key
	}
	cmdArgs = append(cmdArgs, args...)
	cmd := redis.NewCmd(ctx, cmdArgs...)

	if len(keys) > 0 {
		cmd.SetFirstKeyPos(3)
	}

	return cmd
}

// parseBarrierResult parses the barrier result to shield the user-level
// code from perceiving the barrier logic.
func (h *Hook) parseBarrierResult(cmd *redis.Cmd) error {
	// only barrier logic returns single result
	_, ok := cmd.Val().([]any)
	if ok {
		return nil
	}

	brrRet, _ := cmd.Val().(string) // must be string type.
	switch brrRet {
	case barrierResultDuplicate:
		return txbarrier.ErrDuplicationOrSuspension
	case barrierResultEmptyCompensation:
		return txbarrier.ErrEmptyCompensation
	default:
		// never happened
		return nil
	}
}
