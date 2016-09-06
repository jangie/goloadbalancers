package util

import (
	"net/http"
	"net/url"
)

//LoadBalancer must support http.Handler and also give ability to add and remove
type LoadBalancer interface {
	http.Handler
	Add(u *url.URL) (err error)
	Remove(u *url.URL) (err error)
}
