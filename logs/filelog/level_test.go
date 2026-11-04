package filelog

import (
	"bytes"
	"strings"
	"testing"

	"github.com/magic-lib/go-plat-utils/logs"
	"github.com/sirupsen/logrus"
)

func TestFileLoggerLevelFilter(t *testing.T) {
	var buf bytes.Buffer
	oldOut, oldLevel := logrus.StandardLogger().Out, logrus.GetLevel()
	logrus.SetOutput(&buf)
	defer func() {
		logrus.SetOutput(oldOut)
		logrus.SetLevel(oldLevel)
	}()

	x := new(FileLogger)
	x.SetLevel(logs.INFO)
	x.Debug("msg-debug")
	x.Info("msg-info")
	x.Warn("msg-warn")
	x.Error("msg-error")

	out := buf.String()
	for _, want := range []string{"msg-info", "msg-warn", "msg-error"} {
		if !strings.Contains(out, want) {
			t.Errorf("级别为 INFO 时应输出 %s, 实际输出: %q", want, out)
		}
	}
	if strings.Contains(out, "msg-debug") {
		t.Errorf("级别为 INFO 时不应输出 debug 日志, 实际输出: %q", out)
	}
}
