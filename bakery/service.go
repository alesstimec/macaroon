package bakery

// Service represents a service which can delegate
// authorization checks to other services,
// and may also be used as a delegation endpoint
// to discharge authorization checks from other services.
type Service struct {
}

// NewService returns a service which stores its macaroons
// in the given storage. The given checker function is
// used to check the validity of caveats.
func NewService(store Storage, checker FirstPartyChecker) *Service

// AddPublicKeyForLocation specifies that third party caveats
// for the given location will be encrypted with the given public
// key. If prefix is true, any locations with loc as a prefix will
// be also associated with the given key.
func (svc *Service) AddPublicKeyForLocation(loc string, prefix bool, key string)

// Checker returns the checker used by the service.
func (svc *Service) Checker() FirstPartyChecker

// NewRequest returns a new client request object that uses checker to
// verify caveats. If checker is nil, the service's checker will be
// used.
func (svc *Service) NewRequest(checker FirstPartyChecker) *Request

// AddClientMacaroons associates the given macaroons with
// the request. They will be taken into account when req.Check
// is called.
func (req *Request) AddClientMacaroons(ms []*macaroon.Macaroon)

// Caveat represents a condition that must be true for a check to
// complete successfully. If Location is non-empty, the caveat must be
// discharged by a third party at the given location, which should be a
// fully qualified URL that refers to a service which implements the
// name space implemented by Service.DischargeHandler.
type Caveat struct {
	Location  string
	Condition string
}

// Check checks whether the given caveats hold true.
// If they do not hold true because some third party caveat
// is not available, CheckCaveats returns a DischargeRequiredError
// holding a macaroon that must be discharged for the
// given caveats to be fulfilled.
//
// For third-party caveats with no public key
// associated with their location, the requested
// service is contacted and asked to create a new
// caveat identifier.
func (req *Request) Check(caveats []Caveat) error

// DischargeRequiredError represents an error that occurs
// when an operation requires permissions which are not
// available. The Macaroon field holds a macaroon which
// must be discharged for the original operation to succeed.
type DischargeRequiredError struct {
	Macaroon *macaroon.Macaroon
}

// ThirdPartyCaveatId defines the format
// of a third party caveat id. If ThirdPartyPublicKey
// is non-empty, then both FirstPartyPublicKey
// and Nonce must be set, and the id will have
// been encrypted with the third party key.
// If not, the Id holds an id
type ThirdPartyCaveatId struct {
	ThirdPartyPublicKey []byte `json:",omitempty"`
	FirstPartyPublicKey []byte `json:",omitempty"`
	Nonce               []byte `json:",omitempty"`
	Id                  string
}

// ThirdPartyChecker holds a function that checks
// third party caveats for validity. It the
// caveat is valid, it returns a nil error and
// optionally a slice of extra caveats that
// will be added to the discharge macaroon.
type ThirdPartyChecker func(caveat string) ([]Caveat, error)

// FirstPartyChecker holds a function that checks
// first party caveats for validity.
type FirstPartyChecker func() error
