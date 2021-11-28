package subscriber

import (
	"github.com/deckarep/golang-set"
	"testing"
)

func TestUpdateRange(t *testing.T) {

	arr := []int64{1, 2, 3, 4, 5}

	slice := []int64{1, 2}

	result := UpdateRange(arr, slice, func(set mapset.Set, i int64) {
		set.Remove(i)
	})

	t.Logf("Result: %v", result)

	slice = []int64{2}

	result = UpdateRange(result, slice, func(set mapset.Set, i int64) {
		set.Add(i)
	})

	t.Logf("Result: %v", result)
}
