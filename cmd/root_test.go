package cmd

import (
	"testing"
	"time"

	"github.com/fmo/skytui/internal/history"
)

func TestRootCommandVersion(t *testing.T) {
	cmd := newRootCmd(history.Store{}, nil, nil, 25*time.Minute, 5*time.Minute)

	if cmd.Version != "v0.6.0" {
		t.Fatalf("got version %q, want %q", cmd.Version, "v0.6.0")
	}
}
