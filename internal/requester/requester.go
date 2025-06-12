package requester

// Requester takes a request and fans out many instances
// of that request until either the maximum count is reached
// or the duration has been surpassed.
type Requester struct{}

// New instantiates a new instance of Requester and returns
// the ptr to it.
func New() *Requester {
	return &Requester{}
}
