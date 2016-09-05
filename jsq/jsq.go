package jsq

import (
	"fmt"
	"net/http"
	"net/url"
	"sync"
)

//JoinShortestQueueBalancer is a bookkeeping struct
type JoinShortestQueueBalancer struct {
	balancees      map[*url.URL]int
	highWatermark  map[url.URL]int
	requestCounter map[url.URL]int
	isTesting      bool
	next           http.Handler
	keys           []*url.URL
	lock           *sync.Mutex
}

type JoinShortestQueueBalancerOptions struct {
	IsTesting bool
}

func (b *JoinShortestQueueBalancer) nextServer() (*url.URL, error) {
	b.lock.Lock()
	defer b.lock.Unlock()
	//Special case: If balancees are nil or empty, return an error.
	if b.balancees == nil || len(b.balancees) == 0 {
		return nil, fmt.Errorf("Number of balancees is zero, cannot handle")
	}
	//Special case: If balancees is 1, there is no need to balance
	if len(b.balancees) == 1 {
		for key := range b.balancees {
			return key, nil
		}
	}
	var keysCopy = make([]*url.URL, len(b.keys))
	copy(keysCopy, b.keys)

	var bestChoice *url.URL
	var leastConns = -1
	for _, key := range keysCopy {
		if leastConns == -1 {
			leastConns = b.balancees[key]
			bestChoice = key
			continue
		}
		if leastConns > b.balancees[key] {
			leastConns = b.balancees[key]
			bestChoice = key
		}
	}
	return bestChoice, nil
}

//NewJoinShortestQueueBalancer gives a new ChoiceOfBalancer back
func NewJoinShortestQueueBalancer(balancees []string, options JoinShortestQueueBalancerOptions, next http.Handler) *JoinShortestQueueBalancer {
	var b = JoinShortestQueueBalancer{
		lock: &sync.Mutex{},
	}
	b.balancees = make(map[*url.URL]int)
	if options.IsTesting {
		b.requestCounter = make(map[url.URL]int)
		b.highWatermark = make(map[url.URL]int)
	}
	for _, u := range balancees {
		var purl, _ = url.Parse(u)
		b.keys = append(b.keys, purl)
		b.balancees[purl] = 0
	}

	b.next = next
	return &b
}

func (b *JoinShortestQueueBalancer) acquire(u *url.URL) {
	b.lock.Lock()
	defer b.lock.Unlock()
	b.balancees[u]++
	if b.isTesting {
		if b.balancees[u] > b.highWatermark[*u] {
			b.highWatermark[*u] = b.balancees[u]
		}
		b.requestCounter[*u]++
	}
}

func (b *JoinShortestQueueBalancer) release(u *url.URL) {
	b.lock.Lock()
	defer b.lock.Unlock()
	b.balancees[u]--
}

//NumberOfBalancees returns the number of balancees that this balancer knows about
func (b *JoinShortestQueueBalancer) NumberOfBalancees() int {
	return len(b.keys)
}

//OutstandingRequests returns the number of outstanding requests for a particular balancee
func (b *JoinShortestQueueBalancer) OutstandingRequests(u *url.URL) int {
	return b.balancees[u]
}

//HighWatermark returns the most outstanding requests for a particular balancee
func (b *JoinShortestQueueBalancer) HighWatermark(u *url.URL) int {
	return b.highWatermark[*u]
}

//RequestCount gives back the number of requests that have come into a particular URL
func (b *JoinShortestQueueBalancer) RequestCount(u *url.URL) int {
	return b.requestCounter[*u]
}

func (b *JoinShortestQueueBalancer) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if w == nil || req == nil {
		return
	}
	if len(b.keys) == 0 {
		w.WriteHeader(502)
		fmt.Fprint(w, "jsq has no balancees. no backend server available to fulfill this request.")
		return
		//return 502
	}
	var next, _ = b.nextServer()
	newReq := *req
	newReq.URL = next
	b.acquire(next)
	if b.next != nil {
		b.next.ServeHTTP(w, &newReq)
	} else {
		fmt.Fprint(w, "jsq does not have a next middleware and is unable to forward to the balancee.")
	}
	b.release(next)
}
