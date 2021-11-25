package blive

import (
	mapset "github.com/deckarep/golang-set"
	"testing"
)

var a = mapset.NewSet(3, 4, 5, 6)
var b = mapset.NewSet(1, 2, 3, 4)

func TestADiffB(t *testing.T) {
	t.Log(a.Difference(b))
}

func TestBDiffA(t *testing.T) {
	t.Log(b.Difference(a))
}
