package bestof

import (
	"errors"
	"math/rand"
)

//RandomInt gives a nextInt from its range which may or may not actually be random
type RandomInt interface {
	nextInt() (uint64, error)
}

//GoRandom uses Go's internal random generator to return a random number between minimum and maximum
type GoRandom struct {
	minimum, maximum uint64
}

//TestingRandom gives the next value in its array as it's 'nextInt', looping back to the first on reaching the end
type TestingRandom struct {
	values []uint64
	index  int
}

func (g GoRandom) nextInt() (uint64, error) {
	if g.maximum < g.minimum {
		return 0, errors.New("Illegal state: Minimum is greater than maximum")
	}
	var n = g.maximum - g.minimum
	return uint64(rand.Int63n(int64(n))) + g.minimum, nil
}

func (t *TestingRandom) nextInt() (uint64, error) {
	if len(t.values) == 0 {
		return 0, nil
	}

	if t.index > len(t.values)-1 {
		t.index = 0
	}

	var value = t.values[t.index]
	t.index++

	return value, nil
}
