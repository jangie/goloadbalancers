package random

import (
	"fmt"
	"net/http"
	"net/url"
	"reflect"

	"github.com/jangie/goloadbalancers/util"
)

type RandomBalancer struct {
	randomGenerator util.RandomInt
	balancees       []*url.URL
	next            http.Handler
	requestCounter  map[url.URL]int
	isTesting       bool
}

type RandomBalancerOptions struct {
	RandomGenerator util.RandomInt
	IsTesting       bool
}

func (b *RandomBalancer) nextServer() (*url.URL, error) {
	if b.balancees == nil || len(b.balancees) == 0 {
		return nil, fmt.Errorf("Number of balancees is zero, cannot handle")
	}
	//Special case: If balancees is 1, there is no need to balance
	if len(b.balancees) == 1 {
		if b.isTesting {
			b.requestCounter[*b.balancees[0]]++
		}
		return b.balancees[0], nil
	}
	var nextIndex, _ = b.randomGenerator.NextInt(0, len(b.balancees))
	if b.isTesting {
		b.requestCounter[*b.balancees[nextIndex]]++
	}
	return b.balancees[nextIndex], nil
}

//NewRandomBalancer gives a new ChoiceOfBalancer back
func NewRandomBalancer(balancees []string, options RandomBalancerOptions, next http.Handler) *RandomBalancer {
	var b = RandomBalancer{}
	b.balancees = make([]*url.URL, len(balancees))
	if options.IsTesting {
		b.isTesting = true
		b.requestCounter = make(map[url.URL]int)
	}

	for index, u := range balancees {
		var purl, _ = url.Parse(u)
		b.balancees[index] = purl
	}
	if options.RandomGenerator == nil {
		b.randomGenerator = &util.GoRandom{}
	} else {
		b.randomGenerator = options.RandomGenerator
	}
	b.next = next
	return &b
}

//NumberOfBalancees returns the number of balancees that this balancer knows about
func (b *RandomBalancer) NumberOfBalancees() int {
	return len(b.balancees)
}

//RequestCount gives back the number of requests that have come into a particular URL
func (b *RandomBalancer) RequestCount(u *url.URL) int {
	return b.requestCounter[*u]
}

//ConfiguredRandomInt returns the string representation of the random generator assigned to the balancee. Used for testing.
func (b *RandomBalancer) ConfiguredRandomInt() string {
	return reflect.TypeOf(b.randomGenerator).String()
}

func (b *RandomBalancer) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if w == nil || req == nil {
		return
	}
	if len(b.balancees) == 0 {
		w.WriteHeader(502)
		fmt.Fprint(w, "randomlb has no balancees. no backend server available to fulfill this request.")
		return
		//return 502
	}
	var next, _ = b.nextServer()
	newReq := *req
	newReq.URL = next
	if b.next != nil {
		b.next.ServeHTTP(w, &newReq)
	} else {
		fmt.Fprint(w, "random does not have a next middleware and is unable to forward to the balancee.")
	}
}
