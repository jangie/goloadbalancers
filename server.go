package main

import (
	"fmt"
	"net/http"

	"github.com/jangie/bestofnlb/bestof"
	"github.com/vulcand/oxy/forward"
)

//Test harness
type testHarness struct {
	next http.Handler
}

func (t *testHarness) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if req.URL.Path == "/" {
		fmt.Fprint(w, "This is a dumb server which is meant to be used with the node file testServer.js.\n - Run the node server (which will hold onto :8080)\n - Add pointers into your hosts file for 127.0.0.1 testa, testb, testc\n - Hit localhost:8090/simulateServers, and see which server you get balanced to")
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
			RandomGenerator: bestof.GoRandom{},
		},
		fwd,
	)
	var t = testHarness{
		next: bal,
	}

	http.Handle("/", &t)
	http.ListenAndServe(":8090", nil)
}
