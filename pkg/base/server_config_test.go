package base

import (
	"testing"
)

func TestNewStoreDir(t *testing.T) {
	_, err := NewStoreDir("~/test")
	if err == nil {
		t.Error("Starting with '~' is not accepted")
	}
}
