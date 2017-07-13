package cli

import (
	"testing"

	"github.com/taku-k/polymerase/pkg/utils/leaktest"
	"github.com/taku-k/polymerase/pkg/utils/testutil"
)

func TestStartArgsChecking(t *testing.T) {
	defer leaktest.AfterTest(t)()

	f := startCmd.Flags()

	testCases := []struct {
		args     []string
		expected string
	}{
		{[]string{}, ``},
		{[]string{`--store-dir=~/polymerase-data`}, `store path cannot start with '~'`},
	}
	for i, c := range testCases {
		err := f.Parse(c.args)
		if !testutil.IsError(err, c.expected) {
			t.Errorf("%d: expected %q, but found %v", i, c.expected, err)
		}
	}
}
