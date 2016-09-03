package main

import (
	"net/http"

	"github.com/jangie/bestofnlb/bestof"
)

//Test harness
func main() {
	var bal = bestof.NewBalancer(
		[]string{"http://testa:8080", "http://testb:8080", "http://testc:8080"},
		bestof.GoRandom{},
		2,
	)
	http.HandleFunc("/asdf", bal.ServeHTTP)
	http.ListenAndServe(":8090", nil)
}
