package filterlist

import (
	"strings"
	"testing"

	"github.com/sirupsen/logrus"
	logrustest "github.com/sirupsen/logrus/hooks/test"
)

func verifyLogged(fname string, wantLog []string, hook *logrustest.Hook, t *testing.T) {
	if len(wantLog) != len(hook.Entries) {
		logged := make([]string, 0, len(hook.Entries))
		for _, gotLog := range hook.Entries {
			logged = append(logged, gotLog.Message)
		}
		t.Errorf("%s() logged %d lines , want %d\n%#v", fname, len(hook.Entries), len(wantLog), logged)
	} else {
		for i, gotLog := range hook.Entries {
			msg := strings.TrimSuffix(gotLog.Message, "\n")
			if msg != wantLog[i] {
				t.Errorf("%s() logged\n%#v \nwant\n%#v", fname, msg, wantLog[i])
			}
		}
	}
}

func HadErrorLevel(hook *logrustest.Hook, level logrus.Level) bool {
	for _, gotLog := range hook.Entries {
		if gotLog.Level == level {
			return true
		}
	}
	return false
}
