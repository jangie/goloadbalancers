package util

import "testing"

func TestGoRandomImplements(t *testing.T) {
	var random RandomInt
	random = &GoRandom{}
	var nextInt, _ = random.NextInt(0, 100)
	if nextInt >= 100 {
		t.Fatalf("GoRandom gave unexpected answer")
	}
}

func TestTestingRandom(t *testing.T) {

	var values = []int{1, 3, 5}
	var random = &TestingRandom{
		Values: values,
	}
	var i = 0
	for ; i < 100; i++ {
		var nextInt, _ = random.NextInt(0, 100)
		if nextInt != values[i%len(values)] {
			t.Fatalf("We had an unexpected situation with the Testing Random generator")
		}

	}
	if random.CallCount != 100 {
		t.Fatalf("The testing random can't keep track of how many times it's been called!")
	}

}
