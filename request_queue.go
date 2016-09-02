package bestof

//Request represents an outstanding request
type Request struct{}

//Semaphore holds onto the number of outstanding items in a queue
type Semaphore chan Request

//Add a request counter to the queue of work for a backend
func (s Semaphore) addToQueue() {
	r := Request{}
	s <- r
}

//Remove a request counter from the queue of work for a backend
func (s Semaphore) removeFromQueue() {
	<-s
}

//Check the length of the queue of work for a backend
func (s Semaphore) length() int {
	return s.length()
}
