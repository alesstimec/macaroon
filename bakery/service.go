package bakery

// Service represents a service which can delegate
// authorization checks to other services,
// and may also be used as a delegation endpoint
// to discharge authorization checks from other services.
type Service struct {
	store Storage
	checker FirstPartyChecker
	caveatMaker CaveatMaker

	// mu guards the fields following it.
	mu sync.Mutex

	// thirdParty maps client ids to the third-party
	// caveat ids cached for those clients.
	//
	// TODO(rog) we could potentially use
	// external storage for this so that clients
	// could use cached ids even when going to
	// one of several different servers.
	thirdPartyCaveats map[string] *thirdPartyCaveatCache
}

// NewService returns a service which stores its macaroons
// in the given storage. The given checker function is
// used to check the validity of caveats.
func NewService(store Storage, checker FirstPartyChecker, caveatMaker CaveatMaker) *Service {
	return &Service{
		store: store,
		checker: checker,
		caveatMaker: caveatMaker,
		thirdPartyCaveats: make(map[string] *thirdPartyCaveatCache),
	}
}

// CaveatMaker can create caveat ids for
// third parties. It is left abstract to allow location-dependent
// caveat id creation.
type CaveatMaker interface {
	NewCaveatId(caveat Caveat) (string, error)
}

// Checker returns the checker used by the service.
func (svc *Service) Checker() FirstPartyChecker {
	return svc.checker
}

// Request represents a request made to a service
// by a client. The request may be long-lived. It holds a set
// of macaroons that the client wishes to be taken
// into account.
type Request struct {
	svc *Service
	checker FirstPartyChecker
	macaroons map[string] *macaroon.Macaroon
	clientId string
}

// NewRequest returns a new client request object that uses checker to
// verify caveats. If checker is nil, the service's checker will be
// used. The clientId, if non-empty, will be used to associate
// the request with others with the same id - third party
// caveats will be shared between requests with the same clientId,
// allowing a given client to cache them.
func (svc *Service) NewRequest(clientId string, checker FirstPartyChecker) *Request {
	return &Request{
		svc: svc,
		checker: checker,
		clientId: clientId,
	}
}

// AddClientMacaroons associates the given macaroons with
// the request. They will be taken into account when req.Check
// is called.
func (req *Request) AddClientMacaroon(m *macaroon.Macaroon) {
	// TODO(rog) what should we do if there's more than
	// macaroon with the given id?
	// We could potentially try all macaroons with a given id
	// to find one that passes.
	// For the time being, just arbitrarily discard duplicates.
	req.macaroons[m.Id()] = m
}

// Caveat represents a condition that must be true for a check to
// complete successfully. If Location is non-empty, the caveat must be
// discharged by a third party at the given location, which should be a
// fully qualified URL that refers to a service which implements the
// name space implemented by Service.DischargeHandler.
type Caveat struct {
	Location  string
	Condition string
}

// Check checks whether the given caveats hold true. If they do not hold
// true because some third party caveat is not available, CheckCaveats
// returns a DischargeRequiredError holding a macaroon that must be
// discharged for the given caveats to be fulfilled.
func (req *Request) Check(caveats []Caveat) error {
	// check cheap caveats first.
	sort.Stable(byCaveatCost(caveats))
	for _, cav := range caveats {
		if cav.Location == "" {
			if err := req.svc.checker(cav.Condition); err != nil {
				return fmt.Errorf("caveat %q not satisfied: %v", cav.Condition, err)
			}
		} else {
			if err != req.checkThirdParty(cav); err != nil {
				return fmt.Errorf(" oooooh dear
	}
}

func (req *Request) getCaveat(cav 

// DischargeRequiredError represents an error that occurs
// when an operation requires permissions which are not
// available. The Macaroon field holds a macaroon which
// must be discharged for the original operation to succeed.
type DischargeRequiredError struct {
	Macaroon *macaroon.Macaroon
}

// ThirdPartyChecker holds a function that checks
// third party caveats for validity. It the
// caveat is valid, it returns a nil error and
// optionally a slice of extra caveats that
// will be added to the discharge macaroon.
type ThirdPartyChecker func(caveat string) ([]Caveat, error)

// FirstPartyChecker holds a function that checks
// first party caveats for validity.
type FirstPartyChecker func(caveat string) error

// thirdPartyCaveatCache holds a client-specific cache
// of caveat ids.
// TODO(rog) figure out how to expire entries from
// this cache (perhaps time-expire entries based on
// the expiry date of the macroon that includes the
// third-party caveat)
type thirdPartyCaveatCache struct {
	// caveats maps from caveat to caveat id.
	caveats map[Caveat] string
}

func (c *thirdPartyCaveatCache) get(cav Caveat) string {
	return c.caveats[cav]
}

func (c *thirdPartyCaveatCache) put(cav Caveat, caveatId string) {
	c.caveats[cav] = caveatId
}

type byCaveatCost []Caveat

func (c byCaveatCost) Less(i, j int) bool {
	// TODO(rog) more cost criteria?
	iHasLoc, jHasLoc := c[i].Location != "", c[j].Location != ""
	if iHasLoc != jHasLoc {
		return !iHasLoc
	}
	return false
}

func (c byCaveatCost) Swap(i, j int) {
	c[i], c[j] = c[j], c[i]
}

func (c byCaveatCost) Len() int {
	return len(c)
}