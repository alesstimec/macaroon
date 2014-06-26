package bakery

// Discharge can act as a discharger for third-party caveats.
type Discharger struct {
}

// NewDischarger creates a new Discharger that uses
// the given store for storing macaroons and
// given checker for checking third-party caveats.
func NewDischarger(store Storage, checker ThirdPartyChecker) *Discharger

// Discharge attempts to discharge the macaroon at the given
// location with the given id.
func (d *Discharger) Discharge(location, id string) (*macaroon.Macaroon, error)
