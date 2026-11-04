package mysqllog

import "testing"

func TestIsLogTableName(t *testing.T) {
	cases := map[string]bool{
		"app_log_20250101":        true,
		"app_log_20251231":        true,
		"app_log_2025_summary":    false, // 同前缀的其他表
		"appXlog_20250101":        false, // LIKE 中 _ 通配会匹配到
		"app_log_20251301":        false, // 非法日期
		"app_log_202501011":       false,
		"app_log_":                false,
		"other_app_log_20250101":  false,
		"app_log_20250101_backup": false,
	}
	for name, want := range cases {
		if got := isLogTableName("app_log", name); got != want {
			t.Errorf("isLogTableName(app_log, %s) = %v, want %v", name, got, want)
		}
	}
}

func TestEscapeLike(t *testing.T) {
	if got, want := escapeLike(`app_log%x\y`), `app\_log\%x\\y`; got != want {
		t.Errorf("escapeLike = %s, want %s", got, want)
	}
}
