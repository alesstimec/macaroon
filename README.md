# macaroon
--
    import "github.com/rogpeppe/macaroon"

The macaroon package implements macaroons as described in the paper "Macaroons:
Cookies with Contextual Caveats for Decentralized Authorization in the Cloud"
(http://theory.stanford.edu/~ataly/Papers/macaroons.pdf)

It still in its very early stages, having no support for serialisation and only
rudimentary test coverage.

## Usage

#### type Caveat

```go
type Caveat struct {
}
```

Caveat holds a first person or third party caveat.

#### func (*Caveat) IsThirdParty

```go
func (cav *Caveat) IsThirdParty() bool
```
IsThirdParty reports whether the caveat must be satisfied by some third party
(if not, it's a first person caveat).

#### func (*Caveat) MarshalJSON

```go
func (cav *Caveat) MarshalJSON() ([]byte, error)
```
MarshalJSON implements json.Marshaler.

#### func (*Caveat) UnmarshalJSON

```go
func (cav *Caveat) UnmarshalJSON(jsonData []byte) error
```
unmarshalJSON implements json.Unmarshaler.

#### type Macaroon

```go
type Macaroon struct {
}
```

Macaroon holds a macaroon. See Fig. 7 of
http://theory.stanford.edu/~ataly/Papers/macaroons.pdf for a description of the
data contained within. Macaroons are mutable objects - use Clone as appropriate
to avoid unwanted mutation.

#### func  New

```go
func New(rootKey []byte, id, loc string) *Macaroon
```
New returns a new macaroon with the given root key, identifier and location.

#### func (*Macaroon) AddFirstPartyCaveat

```go
func (m *Macaroon) AddFirstPartyCaveat(caveatId string)
```
AddFirstPartyCaveat adds a caveat that will be verified by the target service.

#### func (*Macaroon) AddThirdPartyCaveat

```go
func (m *Macaroon) AddThirdPartyCaveat(thirdPartySecret []byte, caveat string, loc string) (id string, err error)
```
AddThirdPartyCaveat adds a third-party caveat to the macaroon, using the given
shared secret, caveat and location hint. It returns the caveat id of the third
party macaroon.

#### func (*Macaroon) Bind

```go
func (m *Macaroon) Bind(rootSig []byte)
```
Bind prepares the macaroon for being used to discharge the macaroon with the
given rootSig. This must be used before it is used in the discharges argument to
Verify.

#### func (*Macaroon) Caveats

```go
func (m *Macaroon) Caveats() []Caveat
```
Caveats returns the macaroon's caveats. This method will probably change, and
it's important not to change the returned caveat.

#### func (*Macaroon) Clone

```go
func (m *Macaroon) Clone() *Macaroon
```
Clone returns a copy of the receiving macaroon.

#### func (*Macaroon) Id

```go
func (m *Macaroon) Id() string
```
Id returns the id of the macaroon. This can hold arbitrary information.

#### func (*Macaroon) Location

```go
func (m *Macaroon) Location() string
```
Location returns the macaroon's location hint. This is not verified as part of
the macaroon.

#### func (*Macaroon) MarshalJSON

```go
func (m *Macaroon) MarshalJSON() ([]byte, error)
```
MarshalJSON implements json.Marshaler.

#### func (*Macaroon) Signature

```go
func (m *Macaroon) Signature() []byte
```
Signature returns the macaroon's signature.

#### func (*Macaroon) UnmarshalJSON

```go
func (m *Macaroon) UnmarshalJSON(jsonData []byte) error
```
UnmarshalJSON implements json.Unmarshaler.

#### func (*Macaroon) Verify

```go
func (m *Macaroon) Verify(rootKey []byte, check func(caveat string) (bool, error), discharges map[string]*Macaroon) (bool, error)
```
Verify verifies that the receiving macaroon is valid. The root key must be the
same that the macaroon was originally minted with. The check function is called
to verify each first-party caveat - it may return an error the check cannot be
made but the answer is not necessarily false. The discharge macaroons should be
passed in discharges, keyed by macaroon id.

Verify returns true if the verification succeeds; if returns (false, nil) if the
verification fails, and (false, err) if the verification cannot be asserted (but
may not be false).

#### type ThirdPartyCaveatId

```go
type ThirdPartyCaveatId struct {
	RootKey []byte
	Caveat  string
}
```

ThirdPartyCaveatId holds the information encoded in a third-party caveat id.

#### func  DecryptThirdPartyCaveatId

```go
func DecryptThirdPartyCaveatId(secret []byte, id string) (*ThirdPartyCaveatId, error)
```
DecryptThirdPartyCaveatId decrypts a third-party caveat id given the shared
secret.

#### type Verifier

```go
type Verifier interface {
	Verify(m *Macaroon, rootKey []byte) (bool, error)
}
```
# bakery
--
    import "github.com/rogpeppe/macaroon/bakery"


## Usage

```go
var ErrCaveatNotRecognized = "caveat not recogniz.ed"
```

```go
var ErrNotFound = errors.New("item not found")
```

#### type Capability

```go
type Capability struct {
	// Id holds the capability identifier. This
	// should describe the capability in question.
	Id string

	// Caveats holds the list of caveats that must
	// hold for the capability to be granted.
	Caveats []Caveat
}
```

Capability represents a client capability. A client can gain a capability by
presenting a valid, fully discharged macaroon that is associated with the
capability.

#### type Caveat

```go
type Caveat struct {
	Location  string
	Condition string
}
```

Caveat represents a condition that must be true for a check to complete
successfully. If Location is non-empty, the caveat must be discharged by a third
party at the given location, which should be a fully qualified URL that refers
to a service which implements the name space implemented by
Service.DischargeHandler.

#### type CaveatIdMaker

```go
type CaveatIdMaker interface {
	NewCaveatId(caveat Caveat, secret []byte) (string, error)
}
```

CaveatIdMaker can create caveat ids for third parties. It is left abstract to
allow location-dependent caveat id creation.

#### type DischargeRequiredError

```go
type DischargeRequiredError struct {
	Macaroon *macaroon.Macaroon
}
```

DischargeRequiredError represents an error that occurs when an operation
requires permissions which are not available. The Macaroon field holds a
macaroon which must be discharged for the original operation to succeed.

#### type Discharger

```go
type Discharger struct {
}
```

Discharge can act as a discharger for third-party caveats.

#### func  NewDischarger

```go
func NewDischarger(store Storage, checker ThirdPartyChecker) *Discharger
```
NewDischarger creates a new Discharger that uses the given store for storing
macaroons and given checker for checking third-party caveats.

#### func (*Discharger) Discharge

```go
func (d *Discharger) Discharge(location, id string) (*macaroon.Macaroon, error)
```
Discharge attempts to discharge the macaroon at the given location with the
given id.

#### type FirstPartyChecker

```go
type FirstPartyChecker interface {
	CheckFirstPartyCaveat(caveat string) error
}
```

FirstPartyChecker holds a function that checks first party caveats for validity.

If the caveat kind was not recognised, the checker should return
ErrCaveatNotRecognised.

#### type FirstPartyCheckerFunc

```go
type FirstPartyCheckerFunc func(caveat string) error
```


#### func (FirstPartyCheckerFunc) CheckFirstPartyCaveat

```go
func (c FirstPartyCheckerFunc) CheckFirstPartyCaveat(caveat string) error
```

#### type Request

```go
type Request struct {
}
```

Request represents a request made to a service by a client. The request may be
long-lived. It holds a set of macaroons that the client wishes to be taken into
account.

Methods on a Request may be called concurrently with each other.

#### func (*Request) AddClientMacaroon

```go
func (req *Request) AddClientMacaroon(m *macaroon.Macaroon)
```
AddClientMacaroon associates the given macaroon with the request. The macaroon
will be taken into account when req.Check is called.

#### func (*Request) Check

```go
func (req *Request) Check(capability *Capability) error
```
Check checks whether the given caveats hold true. If they do not hold true
because some third party caveat is not available, CheckCaveats returns a
DischargeRequiredError holding a macaroon that must be discharged for the given
caveats to be fulfilled.

#### type Service

```go
type Service struct {
}
```

Service represents a service which can delegate authorization checks to other
services, and may also be used as a delegation endpoint to discharge
authorization checks from other services.

#### func  NewService

```go
func NewService(
	location string,
	store Storage,
	checker FirstPartyChecker,
	caveatIdMaker CaveatIdMaker,
) *Service
```
NewService returns a service which stores its macaroons in the given storage.
The given checker function is used to check the validity of caveats. Macaroons
generated by the service will be associated with the given location.

#### func (*Service) Checker

```go
func (svc *Service) Checker() FirstPartyChecker
```
Checker returns the checker used by the service.

#### func (*Service) NewRequest

```go
func (svc *Service) NewRequest(clientId string, checker FirstPartyChecker) *Request
```
NewRequest returns a new client request object that uses checker to verify
caveats. If checker is nil, the service's checker will be used. The clientId, if
non-empty, will be used to associate the request with others with the same id -
third party caveats will be shared between requests with the same clientId,
allowing a given client to cache them.

#### type Storage

```go
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
```

Storage defines storage for macaroons. TODO(rog) define whether these methods
must be thread-safe or not.

#### func  NewMemStorage

```go
func NewMemStorage() Storage
```
NewMemStorage returns an implementation of Storage that stores all items in
memory.

#### type ThirdPartyChecker

```go
type ThirdPartyChecker interface {
	CheckThirdPartyCaveat(caveat string) ([]Caveat, error)
}
```

ThirdPartyChecker holds a function that checks third party caveats for validity.
It the caveat is valid, it returns a nil error and optionally a slice of extra
caveats that will be added to the discharge macaroon.

If the caveat kind was not recognised, the checker should return
ErrCaveatNotRecognised.
# checkers
--
    import "github.com/rogpeppe/macaroon/bakery/checkers"


## Usage

```go
var Std = Map{
	"expires-before": bakery.FirstPartyCheckerFunc(expiresBefore),
}
```

#### func  ExpiresBefore

```go
func ExpiresBefore(t time.Time) bakery.Caveat
```

#### func  FirstParty

```go
func FirstParty(identifier string, args ...interface{}) bakery.Caveat
```

#### func  ParseCaveat

```go
func ParseCaveat(cav string) (string, string, error)
```
ParseCaveat parses a caveat into an identifier, identifying the checker that
should be used, and the argument to the checker (the rest of the string).

The identifier is taken from all the characters before the first space
character.

#### func  PushFirstPartyChecker

```go
func PushFirstPartyChecker(c0, c1 bakery.FirstPartyChecker) bakery.FirstPartyChecker
```
PushFirstPartyChecker returns a checker that first uses c0 to check caveats, and
falls back to using c1 if c0 returns bakery.ErrCaveatNotRecognized.

#### func  ThirdParty

```go
func ThirdParty(location, identifier string, args ...interface{}) bakery.Caveat
```

#### type Map

```go
type Map map[string]bakery.FirstPartyChecker
```


#### func (Map) CheckFirstPartyCaveat

```go
func (m Map) CheckFirstPartyCaveat(cav string) error
```

#### type StructuredCaveat

```go
type StructuredCaveat struct {
	Identifier string
	Args       []interface{}
}
```
# example
--
# httpbakery
--
    import "github.com/rogpeppe/macaroon/httpbakery"


## Usage

#### func  DischargeHandler

```go
func DischargeHandler(store bakery.Storage, checker bakery.ThirdPartyChecker, key *KeyPair) http.Handler
```
DischargeHandler returns an HTTP handler that issues discharge macaroons to
clients after using the given check function to ensure that the given client is
allowed to obtain a discharge.

The check function is used to check whether a client making the given request
should be allowed a discharge for the given caveat. If it does not return an
error, the caveat will be discharged, with any returned caveats also added to
the discharge macaroon.

If key is not nil, it will be used to decrypt caveat ids, and the public part of
it will be served from /publickey.

The name space served by DischargeHandler is as follows. All parameters can be
provided either as URL attributes or form attributes. The result is always
formatted as a JSON object.

POST /discharge

    params:
    	id: id of macaroon to discharge
    	location: location of original macaroon (optional)
    result:
    	{
    		Macaroon: macaroon in json format
    		Error: string
    	}

POST /create

    params:
    	caveat: caveat to discharge
    	targetLocation: location of target service
    result:
    	{
    		CaveatID: string
    		Error: string
    	}

GET /publickey

    result:
    	public key of service
    	expiry time of key

#### func  Do

```go
func Do(c *http.Client, req *http.Request) (*http.Response, error)
```
Do makes an http request to the given client. If the request fails with a
discharge-required error, any required discharge macaroons will be acquired, and
the request will be repeated with those attached.

If c.Jar field is non-nil, the macaroons will be stored there and made available
to subsequent requests.

#### func  NewHandler

```go
func NewHandler(svc *bakery.Service, handler http.Handler) http.Handler
```
NewHandler returns an http handler that wraps the given handler by creating a
Request for each http.Request that can be retrieved by calling GetRequest.

#### type BakeryRequest

```go
type BakeryRequest struct {
	*bakery.Request
}
```

BakeryRequest wraps *bakery.Request. It is defined to avoid a field clash in the
definition of Request.

#### type CaveatIdMaker

```go
type CaveatIdMaker struct {
}
```

CaveatIdMaker implements bakery.CaveatIdMaker. It knows how to make caveat ids
by communicating with the caveat id creation service served by DischargeHandler,
and also how to create caveat ids using public key cryptography (also recognised
by the DischargeHandler service).

#### func  NewCaveatIdMaker

```go
func NewCaveatIdMaker(key *KeyPair) (*CaveatIdMaker, error)
```
NewCaveatIdMaker returns a new CaveatIdMaker key, which should have been created
using the NACL box.GenerateKey function. The keys may be nil, in which case new
keys will be generated automatically.

#### func (*CaveatIdMaker) AddPublicKeyForLocation

```go
func (m *CaveatIdMaker) AddPublicKeyForLocation(loc string, prefix bool, key *[32]byte)
```
AddPublicKeyForLocation specifies that third party caveats for the given
location will be encrypted with the given public key. If prefix is true, any
locations with loc as a prefix will be also associated with the given key. The
longest prefix match will be chosen. TODO(rog) perhaps string might be a better
representation of public keys?

#### func (*CaveatIdMaker) NewCaveatId

```go
func (m *CaveatIdMaker) NewCaveatId(cav bakery.Caveat, secret []byte) (string, error)
```
NewCaveatId implements bakery.CaveatIdMaker.NewCaveatId.

#### type FirstPartyCaveat

```go
type FirstPartyCaveat func(req *http.Request, caveat string) error
```


#### type KeyPair

```go
type KeyPair struct {
}
```


#### func  GenerateKey

```go
func GenerateKey() (*KeyPair, error)
```

#### type Request

```go
type Request struct {
	*http.Request
	BakeryRequest
}
```

Request holds a request invoked through a handler returned by NewHandler. It
wraps the original http request and the associated bakery request.

#### func  GetRequest

```go
func GetRequest(req *http.Request) *Request
```
GetRequest retrieves the request for the given http request, which must have be
a currently outstanding request invoked through a handler returned by
NewHandler. It panics if there is no associated request.

#### type ThirdPartyCaveat

```go
type ThirdPartyCaveat func(req *http.Request, caveat string) ([]bakery.Caveat, error)
```


#### type ThirdPartyCaveatId

```go
type ThirdPartyCaveatId struct {
	ThirdPartyPublicKey []byte `json:",omitempty"`
	FirstPartyPublicKey []byte `json:",omitempty"`
	Nonce               []byte `json:",omitempty"`
	Id                  []byte
}
```

ThirdPartyCaveatId defines the format of a third party caveat id. If
ThirdPartyPublicKey is non-empty, then both FirstPartyPublicKey and Nonce must
be set, and the id will have been encrypted with the third party key. If not,
the Id holds an id
