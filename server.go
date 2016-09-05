package main

import (
	"fmt"
	"net/http"
	"net/url"

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

func getBestOfHarness(balancees []string, fwd http.Handler) *testHarness {
	var bal = bestof.NewChoiceOfBalancer(
		balancees,
		bestof.ChoiceOfBalancerOptions{
			Choices: 2,
		},
		fwd,
	)
	var tbestof = testHarness{
		next: bal,
		port: 8090,
	}
	return &tbestof
}

func getRandomHarness(balancees []string, fwd http.Handler) *testHarness {
	var random = random.NewRandomBalancer(balancees,
		random.RandomBalancerOptions{
			RandomGenerator: util.GoRandom{},
		},
		fwd,
	)
	var trandom = testHarness{
		next: random,
		port: 8091,
	}
	return &trandom
}

func getRoundRobinHarness(balancees []string, fwd http.Handler) *testHarness {
	var rr, _ = roundrobin.New(fwd)
	var trr = testHarness{
		next: rr,
		port: 8095,
	}
	for _, u := range balancees {
		var purl, _ = url.Parse(u)
		rr.UpsertServer(purl)
	}
	return &trr
}

func getDynamicRoundRobinHarness(balancees []string, fwd http.Handler) *testHarness {
	var rr, _ = roundrobin.New(fwd)
	rebalancer, _ := roundrobin.NewRebalancer(rr)
	for _, u := range balancees {
		var purl, _ = url.Parse(u)
		rr.UpsertServer(purl, roundrobin.Weight(5))
	}
	var tdrr = testHarness{
		next: rebalancer,
		port: 8096,
	}
	return &tdrr
}

func getJSQHarness(balancees []string, fwd http.Handler) *testHarness {
	var jsq = jsq.NewJoinShortestQueueBalancer(balancees,
		jsq.JoinShortestQueueBalancerOptions{},
		fwd,
	)
	var tjsq = testHarness{
		next: jsq,
		port: 8092,
	}
	return &tjsq
}

func main() {
	var fwd, _ = forward.New()
	var balancees = []string{"http://testa:8080", "http://testb:8080", "http://testc:8080"}

	go http.ListenAndServe(":8090", getBestOfHarness(balancees, fwd))
	go http.ListenAndServe(":8091", getRandomHarness(balancees, fwd))
	go http.ListenAndServe(":8092", getJSQHarness(balancees, fwd))

	go http.ListenAndServe(":8095", getRoundRobinHarness(balancees, fwd))
	go http.ListenAndServe(":8096", getDynamicRoundRobinHarness(balancees, fwd))
	fmt.Print("Listening on:\n - http://localhost:8090 [bestof lb]\n - http://localhost:8091 [random lb]\n - http://localhost:8092 [jsq lb]\n - http://localhost:8095 [vulcand/oxy (external) roundrobin]\n - http://localhost:8096 [vulcand/oxy (external) dynamic roundrobin]\n")
	select {}
}
