package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/jangie/goloadbalancers/bestof"
	"github.com/jangie/goloadbalancers/random"
	"github.com/jangie/goloadbalancers/util"
	"github.com/vulcand/oxy/forward"
)

//Test harness
type testHarness struct {
	next http.Handler
	port int
}

func (t *testHarness) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if req.URL.Path == "/" {
		fmt.Fprintf(w, "This is a dumb server which is meant to be used with the node file testServer.js.\n - Run the node server (which will hold onto :8080)\n - Add pointers into your hosts file for 127.0.0.1 testa, testb, testc\n - Hit localhost:%d/simulateServers, and see which server you get balanced to", t.port)
	} else {
		t.next.ServeHTTP(w, req)
	}
}

func main() {
	var fwd, _ = forward.New()
	var bal = bestof.NewChoiceOfBalancer(
		[]string{"http://testa:8080", "http://testb:8080", "http://testc:8080"},
		bestof.ChoiceOfBalancerOptions{
			Choices:         2,
			RandomGenerator: util.GoRandom{},
		},
		fwd,
	)
	var tbestof = testHarness{
		next: bal,
		port: 8090,
	}

	var random = random.NewRandomBalancer([]string{"http://testa:8080", "http://testb:8080", "http://testc:8080"},
		random.RandomBalancerOptions{
			RandomGenerator: util.GoRandom{},
		},
		fwd,
	)
	var trandom = testHarness{
		next: random,
		port: 8091,
	}
	go http.ListenAndServe(":8090", &tbestof)
	go http.ListenAndServe(":8091", &trandom)
	fmt.Print("Listening on http://localhost:8090 [bestof lb] and http://localhost:8091 [random lb]\n\n")
	for true == true {
		time.Sleep(1000)
	}
}
