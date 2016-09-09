package main

import (
	"fmt"
	"net/http"
	"net/url"

	_ "net/http/pprof"

	"github.com/jangie/goloadbalancers/bestof"
	"github.com/jangie/goloadbalancers/jsq"
	"github.com/jangie/goloadbalancers/random"
	"github.com/jangie/goloadbalancers/util"
	"github.com/vulcand/oxy/forward"
	"github.com/vulcand/oxy/roundrobin"
)

//Test harness
type testHarness struct {
	next http.Handler
	port int
}

func (t *testHarness) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if req.URL.Path == "/" {
		fmt.Fprintf(w, "This is a dumb server which is meant to be used with the node file testServer.js.\n - Run the node server (which will hold onto :8080)\n - Add pointers into your hosts file for 127.0.0.1 testa, testb, testc\n - Hit localhost:%d/simulateUnevenServers or localhost:%d/simulateServers, and see which server you get balanced to", t.port, t.port)
	} else {
		t.next.ServeHTTP(w, req)
	}
}

func getBestOfHarness(balancees []url.URL, fwd http.Handler) *testHarness {
	var bal = bestof.NewChoiceOfBalancer(
		balancees,
		bestof.ChoiceOfBalancerOptions{
			Choices: 2,
		},
		fwd,
	)
	return &testHarness{
		next: bal,
		port: 8090,
	}
}

func getRandomHarness(balancees []url.URL, fwd http.Handler) *testHarness {
	var random = random.NewRandomBalancer(balancees,
		random.RandomBalancerOptions{
			RandomGenerator: util.GoRandom{},
		},
		fwd,
	)
	return &testHarness{
		next: random,
		port: 8091,
	}
}

func getRoundRobinHarness(balancees []url.URL, fwd http.Handler) *testHarness {
	var rr, _ = roundrobin.New(fwd)
	for _, u := range balancees {
		rr.UpsertServer(&u)
	}
	var trr = testHarness{
		next: rr,
		port: 8095,
	}
	return &trr
}

func getDynamicRoundRobinHarness(balancees []url.URL, fwd http.Handler) *testHarness {
	var rr, _ = roundrobin.New(fwd)
	rebalancer, _ := roundrobin.NewRebalancer(rr)
	for _, u := range balancees {
		rr.UpsertServer(&u, roundrobin.Weight(5))
	}
	return &testHarness{
		next: rebalancer,
		port: 8096,
	}
}

func getJSQHarness(balancees []url.URL, fwd http.Handler) *testHarness {
	var jsq = jsq.NewJoinShortestQueueBalancer(balancees,
		jsq.JoinShortestQueueBalancerOptions{},
		fwd,
	)
	return &testHarness{
		next: jsq,
		port: 8092,
	}
}

func main() {
	var fwd, _ = forward.New()
	var balanceesStrings = []string{"http://testa:8080", "http://testb:8080", "http://testc:8080"}
	var balancees = []url.URL{}
	for _, u := range balanceesStrings {
		var purl, _ = url.Parse(u)
		balancees = append(balancees, *purl)
	}
	//serve stats for profiling
	go http.ListenAndServe(":8100", http.DefaultServeMux)

	go http.ListenAndServe(":8090", getBestOfHarness(balancees, fwd))
	go http.ListenAndServe(":8091", getRandomHarness(balancees, fwd))
	go http.ListenAndServe(":8092", getJSQHarness(balancees, fwd))

	go http.ListenAndServe(":8095", getRoundRobinHarness(balancees, fwd))
	go http.ListenAndServe(":8096", getDynamicRoundRobinHarness(balancees, fwd))
	fmt.Print("Listening on:\n - http://localhost:8090 [bestof lb]\n - http://localhost:8091 [random lb]\n - http://localhost:8092 [jsq lb]\n - http://localhost:8095 [vulcand/oxy (external) roundrobin]\n - http://localhost:8096 [vulcand/oxy (external) dynamic roundrobin]\n")
	select {}
}
