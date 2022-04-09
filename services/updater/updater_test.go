package updater

import (
	"testing"
)

func TestCheckForUpdates(t *testing.T) {
	if _, err := checkForUpdates(); err != nil {
		t.Fatal(err)
	}
}
