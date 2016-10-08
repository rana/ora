package ora

import "testing"

func TestBoundingPower(t *testing.T) {
	for i, inOut := range [][2]int{
		{0, 0},
		{1, 0},
		{2, 1},
		{3, 2},
		{4, 2},
		{5, 3},
		{7, 3},
		{8, 3},
		{1024, 10},
		{1025, 11},
	} {
		got := boundingPower(inOut[0])
		if got != inOut[1] {
			t.Errorf("%d. (%d) got %d, wanted %d.", i, inOut[0], got, inOut[1])
		}
	}
}
