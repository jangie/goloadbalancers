package bestof

type request struct{}
type semaphore chan request

//Add a request counter to the queue of work for a backend
func (s semaphore) addToQueue() {
	r := request{}
	s <- r
}

//Remove a request counter from the queue of work for a backend
func (s semaphore) removeFromQueue() {
	<-s
}

//Check the length of the queue of work for a backend
func (s semaphore) length() int {
	return s.length()
}
