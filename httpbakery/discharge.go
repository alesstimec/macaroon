package httpbakery

type discharger struct {
	store   bakery.Storage
	checker bakery.ThirdPartyChecker
	key     *KeyPair
}

// DischargeHandler returns an HTTP handler that issues discharge macaroons
// to clients after using the given check function to ensure that the given
// client is allowed to obtain a discharge.
//
// The check function is used to check whether a client making the given
// request should be allowed a discharge for the given caveat. If it
// does not return an error, the caveat will be discharged, with any
// returned caveats also added to the discharge macaroon.
//
// If key is not nil, it will be used to decrypt caveat ids, and
// the public part of it will be served from /publickey.
//
// The name space served by DischargeHandler is as follows.
// All parameters can be provided either as URL attributes
// or form attributes. The result is always formatted as a JSON
// object.
//
// POST /discharge
//	params:
//		id: id of macaroon to discharge
//		location: location of original macaroon (optional)
//	result:
//		{
//			Macaroon: macaroon in json format
//			Error: string
//		}
//
// POST /create
//	params:
//		caveat: caveat to discharge
//		targetLocation: location of target service
//	result:
//		{
//			CaveatID: string
//			Error: string
//		}
//
// GET /publickey
//	result:
//		public key of service
//		expiry time of key
func DischargeHandler(store bakery.Storage, checker bakery.ThirdPartyChecker, key *KeyPair) http.Handler {
	d := &discharger{
		store:   store,
		checker: checker,
		key:     key,
	}
	mux := http.NewServeMux()
	mux.HandleFunc("discharge", d.discharge)
	mux.HandleFunc("create", d.create)
	if key != nil {
		mux.HandleFunc("publickey", d.publicKey)
	}
	return mux
}

func (d *discharger) discharge(w http.ResponseWriter, req *http.Request) {
	if err := req.ParseForm(); err != nil {
		http.Error(fmt.Sprintf("cannot parse form: %v", err), http.StatusBadRequest)
	}
	id := req.Form.Get("id")
	location := req.Form.Get("location")
}

func (d *discharger) create(w http.ResponseWriter, r *http.Request) {
}

func (d *discharger) publickey(w http.ResponseWriter, r *http.Request) {
	// TODO(rog) implement this
	http.Error(w, "not implemented yet", http.StatusNotImplemented)
}
