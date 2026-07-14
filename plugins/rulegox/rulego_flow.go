package rulegox

import (
	"context"
	"fmt"
	"github.com/magic-lib/go-plat-utils/conv"
	"github.com/magic-lib/go-plat-utils/plugins/paramx"
	"github.com/rulego/rulego"
	"github.com/rulego/rulego/api/types"
)

type ActivityFlowConfig struct {
	ChainId     string
	ChainConfig string
	Variables   map[string]any
	MsgType     string
	IsAsync     bool
	EndFunc     func(ctx context.Context, param *paramx.ParamCtx, err error)
}

func StartActivityFlow(actConfig *ActivityFlowConfig) error {
	if actConfig == nil {
		return fmt.Errorf("参数不能为空")
	}
	if actConfig.EndFunc == nil {
		actConfig.EndFunc = func(ctx context.Context, param *paramx.ParamCtx, err error) {
			if err != nil {
				fmt.Printf("工作流执行失败: %v\n", err)
				return
			}
			fmt.Printf("工作流执行成功: %v\n", param)
		}
	}

	engine, err := rulego.New(actConfig.ChainId, []byte(actConfig.ChainConfig))
	if err != nil {
		return err
	}
	paramInput := paramx.NewParamCtxFromVariables(actConfig.Variables)

	if actConfig.MsgType == "" {
		actConfig.MsgType = "ACTIVITY_EVENT"
	}

	msg := types.NewMsg(0, actConfig.MsgType, types.JSON, types.NewMetadata(), conv.String(paramInput))
	endOption := types.WithOnEnd(func(ctx types.RuleContext, msg types.RuleMsg, err error, relationType string) {
		if err != nil {
			actConfig.EndFunc(ctx.GetContext(), nil, err)
			return
		}
		var resultParam = new(paramx.ParamCtx)
		_ = conv.Unmarshal(msg.GetData(), resultParam)

		actConfig.EndFunc(ctx.GetContext(), resultParam, nil)
	})
	if actConfig.IsAsync {
		engine.OnMsg(msg, endOption)
	} else {
		engine.OnMsgAndWait(msg, endOption)
	}
	return nil
}
