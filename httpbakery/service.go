// The httpbakery package layers on top of the bakery
// package - it provides an HTTP-based implementation
// of a macaroon client and server.
package httpbakery
import (
	"fmt"

	"github.com/rogpeppe/macaroon/bakery"
)

// Service represents a service that can use client-provided
// macaroons to authorize requests.
type Service struct {
	*bakery.Service
	caveatIdEncoder *caveatIdEncoder
	key KeyPair
}

// NewServiceParams holds parameters for the NewService call.
type NewServiceParams struct {
	// Location holds the location of the service.
	// Macaroons minted by the service will have this location.
	Location string

	// Store defines where macaroons are stored.
	Store bakery.Storage

	// Key holds the private/public key pair for
	// the service to use. If it is nil, a new key pair
	// will be generated.
	Key *KeyPair
}

// NewService returns a new Service.
func NewService(p NewServiceParams) (*Service, error) {
	if p.Key == nil {
		key, err := GenerateKey()
		if err != nil {
			return nil, fmt.Errorf("cannot generate key: %v", err)
		}
		p.Key = key
	}
	enc := newCaveatIdEncoder(p.Key)
	return &Service{
		Service: bakery.NewService(bakery.NewServiceParams{
			Location: p.Location,
			Store: p.Store,
			CaveatIdEncoder: enc,
		}),
		caveatIdEncoder: enc,
		key: *p.Key,
	}, nil
}

// AddPublicKeyForLocation specifies that third party caveats
// for the given location will be encrypted with the given public
// key. If prefix is true, any locations with loc as a prefix will
// be also associated with the given key. The longest prefix
// match will be chosen.
// TODO(rog) perhaps string might be a better representation
// of public keys?
func (svc *Service) AddPublicKeyForLocation(loc string, prefix bool, publicKey *[32]byte) {
	svc.caveatIdEncoder.addPublicKeyForLocation(loc, prefix, publicKey)
}