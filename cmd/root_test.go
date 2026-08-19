package cmd

import (
	"testing"
	"time"

	"github.com/fmo/skytui/internal/session"
)

func TestRootCommandVersion(t *testing.T) {
	cmd := newRootCmd(session.Store{}, 25*time.Minute, 5*time.Minute)

	if cmd.Version != "v0.5.0" {
		t.Fatalf("got version %q, want %q", cmd.Version, "v0.5.0")
	}
}
