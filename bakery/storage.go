package bakery

// Storage defines storage for macaroons.
type Storage interface {
	// Put stores the item at the given location, overwriting
	// any item that might already be there.
	// TODO(rog) would it be better to lose the overwrite
	// semantics?
	Put(location string, item string) error

	// Get retrieves an item from the given location.
	// If the item is not there, it returns ErrNotFound.
	Get(location string) (item string, err error)

	// Del deletes the item from the given location.
	Del(location string) error
}

var ErrNotFound = errors.New("item not found")

// NewMemStorage returns an implementation of Storage
// that stores all items in memory.
func NewMemStorage() Storage
