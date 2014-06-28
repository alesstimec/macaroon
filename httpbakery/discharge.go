package httpbakery

// TODO(rog) perhaps rename this file to "thirdparty.go" ?

type dischargeHandler struct {
	discharger *baker.Discharger
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
//		location: location of original macaroon (optional (?))
//	result:
//		{
//			Macaroon: macaroon in json format
//			Error: string
//		}
//
// POST /create
//	params:
//		condition: caveat condition to discharge
//		rootkey: root key of discharge caveat
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
func AddDischargeHandler(
		root string,
		mux *http.ServeMux,
		svc *bakery.Service,
		key *KeyPair,
		checker bakery.ThirdPartyChecker,
) {
	d := &dischargeHandler{
		discharger: &bakery.Discharger{
			Checker: checker,
			Decoder: NewCaveatIdDecoder(svc.Store(), key),
			Factory: svc,
		},
		key: key,
	}
	mux.HandleFunc(path.Join(root, "discharge"), d.discharge)
	mux.HandleFunc(path.Join(root, "create"), d.create)
	if key != nil {
		mux.HandleFunc(path.Join(root, "publickey"), d.publicKey)
	}
}

type dischargeResponse struct {
	Macaroon *macaroon.Macaroon
}

func (d *discharger) discharge(w http.ResponseWriter, req *http.Request) {
	if req.Method != "POST" {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	req.ParseForm()
	id := req.Form.Get("id")

	// TODO(rog) pass location into discharge
	// location := req.Form.Get("location")

	m, err := d.discharger.Discharge(id)
	if err != nil {
		http.Error(w, fmt.Sprintf("cannot discharge: %v", err), http.StatusForbidden)
		return
	}
	respBytes, err := json.Marshal(dischargeResponse{m})
	if err != nil {
		d.internalError(w, "cannot marshal response: %v", err)
	}
	w.Write(respBytes)
}

type thirdPartyCaveatIdRecord struct {
	RootKey []byte
	Condition string
}

type createResponse struct {
	CaveatId string
}

func (d *discharger) create(w http.ResponseWriter, r *http.Request) {
	req.ParseForm()
	condition := req.Form.Get("condition")
	rootKeyStr := req.Form.Get("root-key")

	if len(condition) == 0 || len(rootKeyStr) == 0 {
		d.badRequest(w, "empty values for condition or root key")
		return
	}
	rootKey, err := base64.StdEncoding.DecodeFromString(rootKeyStr)
	if err != nil {
		d.badRequest(w, "cannot base64-decode root key: %v", err)
		return
	}
	// TODO(rog) what about expiry times?
	idBytes, err := randomBytes(24)
	if err != nil {
		d.internalError(w, fmt.Errorf("cannot generate random key: %v", err))
		return
	}
	internalId := fmt.Sprintf("third-party-%x", idBytes)
	err := d.store.Put(internalId, thirdPartyCaveatIdRecord{
		Condition: condition,
		RootKey: rootKey,
	})
	if err != nil {
		d.internalError(w, fmt.Errorf("cannot store caveat id record: %v", err))
		return
	}
	idBytes, err := json.Marshal(&ThirdPartyCaveatId{
		Id: internalId,
	})
	if err != nil {
		d.internalError(w, fmt.Errorf("cannot marshal caveat id: %v", err))
		return
	}
	respBytes, err := json.Marshal(createResponse{
		CaveatId: base64.StdEncoding.EncodeToString(idBytes),
	})
	if err != nil {
		d.internalError(w, fmt.Errorf("cannot marshal caveat response: %v", err))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(respBytes)	
}

func (d *discharger) publickey(w http.ResponseWriter, r *http.Request) {
	// TODO(rog) implement this
	http.Error(w, "not implemented yet", http.StatusNotImplemented)
}

func randomBytes(n int) (byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		return fmt.Errorf("cannot generate %d random bytes: %v", n, err)
	}
	return nil
}
