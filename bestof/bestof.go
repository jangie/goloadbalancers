package bestof

import (
	"fmt"
	"net/http"
	"net/url"
	"reflect"
	"sync"
)

//TODO:
// - #semaphore concept: See http://www.golangpatterns.info/concurrency/semaphores
// - #Map of 'backends' to semaphore<empty struct>
// - #Random number provider
// - #Ability to insert random number provider for testing purposes
// - #Code for request proxying
// - #Increment and decrement counter appropriately during proxy
// - #If number of backends is zero, return 502
// - #If number of backends is one, follow next immediately, do not follow algorithm
// - #N should never be zero or greater than number of backends, default to two or number of backends
// - #Configuration should allow for selection of N, should default to two if not present
// - #If N is equal to number of backends, fall back to JSQ, do not go through selection
// - #If N is less than number of backends, randomly select N backends and choose the least loaded

//ChoiceOfBalancer is a bookkeeping struct
type ChoiceOfBalancer struct {
	balancees       map[*url.URL]int
	highWatermark   map[url.URL]int
	requestCounter  map[url.URL]int
	isTesting       bool
	randomGenerator RandomInt
	next            http.Handler
	choices         int
	keys            []*url.URL
	lock            *sync.Mutex
}

type ChoiceOfBalancerOptions struct {
	RandomGenerator RandomInt
	Choices         int
	IsTesting       bool
}

//constructor must:
//- set up keys to be keys of balancees Map
//- set up randomGenerator to choose between 0 and

func (b *ChoiceOfBalancer) nextServer() (*url.URL, error) {
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
	var normalizedChoices = b.choices
	//Special case: If choices is <= 0, default to 2
	if normalizedChoices <= 0 {
		normalizedChoices = 2
	}
	//Special case: If choices > number of balancees, default to number of backends
	if normalizedChoices > len(b.balancees) {
		normalizedChoices = len(b.balancees)
	}
	var potentialChoices = make([]*url.URL, normalizedChoices)
	var keysCopy = make([]*url.URL, len(b.keys))
	copy(keysCopy, b.keys)

	if normalizedChoices == len(b.balancees) {
		potentialChoices = keysCopy
	} else {
		//shuffle keys, we'll choose the first N from the shuffled result
		for i := range keysCopy {
			j, _ := b.randomGenerator.nextInt(0, i+1)
			keysCopy[i], keysCopy[j] = keysCopy[j], keysCopy[i]
		}
		potentialChoices = keysCopy
	}

	var bestChoice *url.URL
	var leastConns = -1
	for index, key := range potentialChoices {
		if index > normalizedChoices {
			break
		}
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

//NewChoiceOfBalancer gives a new ChoiceOfBalancer back
func NewChoiceOfBalancer(balancees []string, options ChoiceOfBalancerOptions, next http.Handler) *ChoiceOfBalancer {
	var b = ChoiceOfBalancer{
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
	if options.RandomGenerator == nil {
		b.randomGenerator = &GoRandom{}
	} else {
		b.randomGenerator = options.RandomGenerator
	}
	if options.Choices == 0 {
		b.choices = 2
	} else {
		b.choices = options.Choices
	}
	b.next = next
	return &b
}

func (b *ChoiceOfBalancer) acquire(u *url.URL) {
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

func (b *ChoiceOfBalancer) release(u *url.URL) {
	b.lock.Lock()
	defer b.lock.Unlock()
	b.balancees[u]--
}

//NumberOfBalancees returns the number of balancees that this balancer knows about
func (b *ChoiceOfBalancer) NumberOfBalancees() int {
	return len(b.keys)
}

//OutstandingRequests returns the number of outstanding requests for a particular balancee
func (b *ChoiceOfBalancer) OutstandingRequests(u *url.URL) int {
	return b.balancees[u]
}

//HighWatermark returns the most outstanding requests for a particular balancee
func (b *ChoiceOfBalancer) HighWatermark(u *url.URL) int {
	return b.highWatermark[*u]
}

func (b *ChoiceOfBalancer) RequestCount(u *url.URL) int {
	return b.requestCounter[*u]
}

//ConfiguredChoices returns the configured number of choices to randomly choose and then pick the best of
func (b *ChoiceOfBalancer) ConfiguredChoices() int {
	return b.choices
}

//ConfiguredRandomInt returns the string representation of the random generator assigned to the balancee. Used for testing.
func (b *ChoiceOfBalancer) ConfiguredRandomInt() string {
	return reflect.TypeOf(b.randomGenerator).String()
}

func (b *ChoiceOfBalancer) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if w == nil || req == nil {
		return
	}
	if len(b.keys) == 0 {
		w.WriteHeader(502)
		fmt.Fprint(w, "bestofnlb has no balancees. no backend server available to fulfill this request.")
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
		fmt.Fprint(w, "bestofnlb does not have a proxy middleware next and is unable to forward to the balancee.")
	}
	b.release(next)
}
