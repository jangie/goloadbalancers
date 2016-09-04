package bestof

import (
	"net/http"
	"net/url"
	"sync"
	"testing"
	"time"

	"github.com/jangie/bestofnlb/util"
)

type testHTTPResponseWriter struct {
	lock         *sync.Mutex
	timesWritten int
	throttler    *testRequestThrottler
}

func (t *testHTTPResponseWriter) Header() http.Header {
	return nil
}

func (t *testHTTPResponseWriter) Write(b []byte) (int, error) {
	if t.throttler != nil {
		t.throttler.releaseRequest()
	}
	t.lock.Lock()
	defer t.lock.Unlock()
	t.timesWritten++
	return 0, nil
}

func (t *testHTTPResponseWriter) WriteHeader(int) {

}

type testHTTPHandler struct{}

func (t *testHTTPHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Host == "a" {
		time.Sleep(time.Duration(10) * time.Millisecond)
	}
	if r.URL.Host == "b" {
		time.Sleep(time.Duration(100) * time.Millisecond)
	}
	if r.URL.Host == "c" {
		time.Sleep(time.Duration(300) * time.Millisecond)
	}
	w.Write([]byte{})
}

type testRequestThrottler struct {
	lock                        *sync.Mutex
	outstandingRequests         int
	maximumSimultaneousRequests int
}

func (t *testRequestThrottler) tryAcquireRequest() bool {
	if t.outstandingRequests < t.maximumSimultaneousRequests {
		t.lock.Lock()
		defer t.lock.Unlock()
		t.outstandingRequests++
		return true
	}
	return false
}

func (t *testRequestThrottler) acquireRequest() {
	for t.tryAcquireRequest() != true {
	}
}

func (t *testRequestThrottler) releaseRequest() {
	t.lock.Lock()
	defer t.lock.Unlock()
	t.outstandingRequests--
}

func TestBestOfImplements(t *testing.T) {
	var handler http.Handler
	handler = NewChoiceOfBalancer([]string{}, ChoiceOfBalancerOptions{}, nil)
	handler.ServeHTTP(nil, nil)
}

func TestBestOfDefaults(t *testing.T) {
	var handler = NewChoiceOfBalancer([]string{}, ChoiceOfBalancerOptions{}, nil)
	if handler.ConfiguredChoices() != 2 {
		t.Fatalf("Configured Choices should default to 2 if not provided")
	}
	if handler.ConfiguredRandomInt() != "*bestof.GoRandom" {
		t.Fatalf("Configured random generator should default to GoRandom if not provided, was %s", handler.ConfiguredRandomInt())
	}
}

func TestBestOfWithOneBalanceeNeverRandomlyChooses(t *testing.T) {
	var randomGenerator = &util.TestingRandom{
		Values: []int{0},
	}
	var handler = NewChoiceOfBalancer([]string{"http://a"}, ChoiceOfBalancerOptions{
		RandomGenerator: randomGenerator,
		IsTesting:       true,
	}, nil)
	handler.ServeHTTP(&testHTTPResponseWriter{lock: &sync.Mutex{}}, &http.Request{})
	if randomGenerator.CallCount > 0 {
		t.Fatalf("Random generator shouldn't be called! The degenerate case of one balancee means that it should be returned.")
	}
}

func TestBestOfWithThreeBalanceesAndTwoChoicesRandomlyChooses(t *testing.T) {
	var randomGenerator = &util.TestingRandom{
		Values: []int{0},
	}
	var handler = NewChoiceOfBalancer([]string{"http://a", "http://b", "http://c"}, ChoiceOfBalancerOptions{
		RandomGenerator: randomGenerator,
		Choices:         2,
		IsTesting:       true,
	}, nil)
	handler.ServeHTTP(&testHTTPResponseWriter{lock: &sync.Mutex{}}, &http.Request{})
	if randomGenerator.CallCount == 0 {
		t.Fatalf("Random generator should've been called! There are three balancees and two choices!")
	}
}

func TestBestOfWithThreeBalanceesAndThreeChoicesDoesNotRandomlyChoose(t *testing.T) {
	var randomGenerator = &util.TestingRandom{
		Values: []int{0},
	}
	var handler = NewChoiceOfBalancer([]string{"http://a", "http://b", "http://c"}, ChoiceOfBalancerOptions{
		RandomGenerator: randomGenerator,
		Choices:         3,
		IsTesting:       true,
	}, nil)
	handler.ServeHTTP(&testHTTPResponseWriter{lock: &sync.Mutex{}}, &http.Request{})
	if randomGenerator.CallCount > 0 {
		t.Fatalf("Random generator shouldn't be called! The degenerate case of #balancees == choices means that we should use JSQ.")
	}
}

// - #Increment and decrement counter appropriately during proxy
// - #If number of backends is zero, return 502
// - #N should never be zero or greater than number of backends, default to two or number of backends
// - #If N is equal to number of backends, fall back to JSQ, do not go through selection
// - #If N is less than number of backends, randomly select N backends and choose the least loaded
func TestPerformsJSQWhenBalanceesEqualNumberOfChoices(t *testing.T) {
	var randomGenerator = &util.GoRandom{}
	var next = &testHTTPHandler{}
	var handler = NewChoiceOfBalancer([]string{"http://a", "http://b", "http://c"}, ChoiceOfBalancerOptions{
		RandomGenerator: randomGenerator,
		Choices:         3,
		IsTesting:       true,
	}, next)
	var i = 0
	var requestThrottler = &testRequestThrottler{lock: &sync.Mutex{}, maximumSimultaneousRequests: 60}
	var writer = &testHTTPResponseWriter{lock: &sync.Mutex{}, throttler: requestThrottler}
	var numberOfWrites = 2000
	for ; i < numberOfWrites; i++ {
		requestThrottler.acquireRequest()
		go handler.ServeHTTP(writer, &http.Request{})
	}
	var allWritten = false
	for allWritten != true {
		time.Sleep(100)
		if writer.timesWritten == numberOfWrites {
			allWritten = true
		}
	}

	var a, _ = url.Parse("http://a")
	var b, _ = url.Parse("http://b")
	var c, _ = url.Parse("http://c")
	if handler.RequestCount(a) < handler.RequestCount(b) || handler.RequestCount(b) < handler.RequestCount(c) {
		t.Fatalf("We are either unlucky or are not following JSQ")
	}
}

func TestPerformsChooseJSQWhenBalanceesGreaterThanNumberOfChoices(t *testing.T) {
	var randomGenerator = &util.TestingRandom{
		Values: []int{0, 1},
	}
	var next = &testHTTPHandler{}
	var handler = NewChoiceOfBalancer([]string{"http://a", "http://b", "http://c"}, ChoiceOfBalancerOptions{
		RandomGenerator: randomGenerator,
		Choices:         2,
		IsTesting:       true,
	}, next)
	var i = 0
	var requestThrottler = &testRequestThrottler{lock: &sync.Mutex{}, maximumSimultaneousRequests: 60}
	var writer = &testHTTPResponseWriter{lock: &sync.Mutex{}, throttler: requestThrottler}
	var numberOfWrites = 2000
	for ; i < numberOfWrites; i++ {
		requestThrottler.acquireRequest()
		go handler.ServeHTTP(writer, &http.Request{})
	}
	var allWritten = false
	for allWritten != true {
		time.Sleep(100)
		if writer.timesWritten == numberOfWrites {
			allWritten = true
		}
	}
	var a, _ = url.Parse("http://a")
	var b, _ = url.Parse("http://b")
	var c, _ = url.Parse("http://c")
	if handler.RequestCount(a) < handler.RequestCount(b) || handler.RequestCount(b) < handler.RequestCount(c) {
		t.Fatalf("We are either unlucky or are not following JSQ")
	}
	if randomGenerator.CallCount == 0 {
		t.Fatalf("We are not shuffling the keys, this means we are not following power of choice")
	}
}
