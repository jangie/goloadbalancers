package bestof

import "testing"

func TestGoRandomImplements(t *testing.T) {
	var random RandomInt
	random = &GoRandom{
		minimum: 0,
		maximum: 100,
	}
	var nextInt, _ = random.nextInt()
	if nextInt >= 100 {
		t.Fatalf("GoRandom gave unexpected answer")
	}
}

func TestTestingRandom(t *testing.T) {
	var random RandomInt
	var values = []uint64{1, 3, 5}
	random = &TestingRandom{
		values: values,
	}
	var i = 0
	for ; i < 100; i++ {
		var nextInt, _ = random.nextInt()
		if nextInt != values[i%len(values)] {
			t.Fatalf("We had an unexpected situation with the Testing Random generator")
		}

	}
}
