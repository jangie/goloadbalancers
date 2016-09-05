package jsq

import (
	"net/http"
	"net/url"
	"sync"
	"testing"
	"time"
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

func TestJSQImplements(t *testing.T) {
	var handler http.Handler
	handler = NewJoinShortestQueueBalancer([]string{}, JoinShortestQueueBalancerOptions{}, nil)
	handler.ServeHTTP(nil, nil)
}

func TestJSQDefaults(t *testing.T) {
	var handler = NewJoinShortestQueueBalancer([]string{}, JoinShortestQueueBalancerOptions{}, nil)
	if handler.isTesting {
		t.Fatalf("Should not default to testing mode")
	}
}

func TestJSQBehavior(t *testing.T) {
	var handler = NewJoinShortestQueueBalancer([]string{"http://a", "http://b", "http://c"}, JoinShortestQueueBalancerOptions{
		IsTesting: true,
	}, nil)

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
