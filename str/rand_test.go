package str

import (
	"testing"
)

func TestRandText(t *testing.T) {
	uniq := make(map[string]bool)
	for i := 0; i < 5000; i++ {
		key := RandText(32)
		if _, ok := uniq[key]; !ok {
			uniq[key] = true
		} else {
			t.Fatalf("RandText uniq fail in round %d", i)
		}
	}
}
