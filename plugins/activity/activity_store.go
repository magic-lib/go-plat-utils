package activity

import "context"

type Store interface {
	// Save 保存（upsert）单个 activity，key 用 act.Type() 的结果。
	Save(ctx context.Context, act *Activity) error
	// List 拉取全部已注册的 activity。
	List(ctx context.Context) ([]*Activity, error)
}
