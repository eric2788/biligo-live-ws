package subscriber

import (
	"testing"
	"time"
)

func BenchmarkAdd(b *testing.B) {

	for i := 0; i < 3773; i++ {
		Add("tester", []int64{int64(i)})
	}

	go func() {
		for true {
			GetAllSubscribers(1)
		}
	}()

	start := time.Now()

	Update("others", []int64{1145141919810})

	elapsed := time.Since(start)
	b.Logf("took %s", elapsed)

}
