package updater

import (
	"fmt"
	"github.com/rogpeppe/go-internal/semver"
	"testing"
)

func TestCheckForUpdates(t *testing.T) {
	if _, err := checkForUpdates(); err != nil {
		t.Fatal(err)
	}
}

func TestCompare(t *testing.T) {
	t.Log(compare("v1.0.0", "v1.0.0"))
	t.Log(compare("v1.0.1", "v1.1.0"))
	t.Log(compare("v0.0.9", "v0.0.10"))
	t.Log(compare("v3.0.1", "v0.9.9"))
	t.Log(compare("v1.0.0", "master"))
}

func compare(v1, v2 string) string {
	return fmt.Sprintf("%v > %v: %v", v1, v2, semver.Compare(v1, v2))
}
