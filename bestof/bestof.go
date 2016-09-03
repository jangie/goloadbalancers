package bestof

import (
	"fmt"
	"net/http"
	"net/url"
	"sync"
)

//TODO:
// - #semaphore concept: See http://www.golangpatterns.info/concurrency/semaphores
// - #Map of 'backends' to semaphore<empty struct>
// - #Random number provider
// - #Ability to insert random number provider for testing purposes
// - #Code for request proxying
// - #Increment and decrement counter appropriately during proxy
// - If number of backends is zero, return 502
// - #If number of backends is one, follow next immediately, do not follow algorithm
// - #N should never be zero or greater than number of backends, default to two or number of backends
// - #Configuration should allow for selection of N, should default to two if not present
// - #If N is equal to number of backends, fall back to JSQ, do not go through selection
// - #If N is less than number of backends, randomly select N backends and choose the least loaded

//Balancer is a bookkeeping struct
type Balancer struct {
	balancees       map[*url.URL]int
	randomGenerator RandomInt
	next            http.Handler
	choices         int
	keys            []*url.URL
	lock            *sync.Mutex
}

//constructor must:
//- set up keys to be keys of balancees Map
//- set up randomGenerator to choose between 0 and

func (b *Balancer) nextServer() (*url.URL, error) {
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
	var potentialChoices = []*url.URL{}

	//shuffle keys, we'll choose the first N from the shuffled result
	for i := range b.keys {
		j, _ := b.randomGenerator.nextInt(0, i+1)
		b.keys[i], b.keys[j] = b.keys[j], b.keys[i]
	}

	if normalizedChoices == len(b.balancees) {
		potentialChoices = b.keys
	} else {
		for i := 0; i < normalizedChoices; i++ {
			potentialChoices = append(potentialChoices, b.keys[i])
		}
	}

	var bestChoice *url.URL
	var leastConns = -1
	fmt.Print(potentialChoices)
	for _, key := range potentialChoices {
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

//NewBalancer gives a new balancer back
func NewBalancer(balancees []string, randomInt RandomInt, choices int, next http.Handler) *Balancer {
	var b = Balancer{
		lock: &sync.Mutex{},
	}
	b.balancees = make(map[*url.URL]int)
	for _, u := range balancees {
		var purl, _ = url.Parse(u)
		b.keys = append(b.keys, purl)
		b.balancees[purl] = 0
	}
	b.randomGenerator = randomInt
	if choices == 0 {
		choices = 2
	}
	b.choices = choices
	b.next = next
	return &b
}

func (b *Balancer) acquire(u *url.URL) {
	b.lock.Lock()
	defer b.lock.Unlock()
	b.balancees[u]++
}

func (b *Balancer) release(u *url.URL) {
	b.lock.Lock()
	defer b.lock.Unlock()
	b.balancees[u]--
}

func (b *Balancer) ServeHTTP(w http.ResponseWriter, req *http.Request) {
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
	b.next.ServeHTTP(w, &newReq)
	b.release(next)
}
