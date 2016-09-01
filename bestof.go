package bestof

//TODO:
// - #semaphore concept: See http://www.golangpatterns.info/concurrency/semaphores
// - Map of 'backends' to semaphore<empty struct>
// - #Random number provider
// - Ability to insert random number provider for testing purposes
// - Code for request proxying
// - If number of backends is zero, throw error
// - If number of backends is one, return immediately, do not follow algorithm
// - N should never be zero or greater than number of backends, default to two or number of backends
// - Configuration should allow for selection of N, should default to two if not present
// - If N is equal to number of backends, fall back to JSQ, do not go through selection
// - If N is less than number of backends, randomly select N backends and choose the least loaded
