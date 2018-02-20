package cli

import (
	"testing"

	"github.com/fortytw2/leaktest"
	"github.com/hatena/polymerase/pkg/utils/testutil"
)

func TestRestoreArgsChecking(t *testing.T) {
	defer leaktest.Check(t)()

	f := restoreCmd.Flags()

	testCases := []struct {
		args     []string
		expected string
	}{
		{[]string{}, ``},
		{[]string{`--max-bandwidth=50mb`}, ``},
		{[]string{`--max-bandwidth=50ab`}, `unhandled size name:`},
		{[]string{`--use-memory=16GB`}, ``},
		{[]string{`--use-memory=15ab`}, `unhandled size name:`},
	}
	for i, c := range testCases {
		err := f.Parse(c.args)
		if !testutil.IsError(err, c.expected) {
			t.Errorf("%d: expected %q, but found %v", i, c.expected, err)
		}
	}
}
