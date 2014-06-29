package httpbakery
import (
	"crypto/rand"
	"fmt"
	"encoding/json"
	"net/http"
	"path"
	"encoding/base64"

	"github.com/rogpeppe/macaroon"
	"github.com/rogpeppe/macaroon/bakery"
)

// TODO(rog) perhaps rename this file to "thirdparty.go" ?

type dischargeHandler struct {
	discharger *bakery.Discharger
	key     *KeyPair
	store bakery.Storage
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
		store: svc.Store(),
	}
	mux.HandleFunc(path.Join(root, "discharge"), d.serveDischarge)
	mux.HandleFunc(path.Join(root, "create"), d.serveCreate)
	if key != nil {
		mux.HandleFunc(path.Join(root, "publickey"), d.servePublicKey)
	}
}

type dischargeResponse struct {
	Macaroon *macaroon.Macaroon
}

func (d *dischargeHandler) serveDischarge(w http.ResponseWriter, req *http.Request) {
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

func (d *dischargeHandler) internalError(w http.ResponseWriter, f string, a ...interface{}) {
	http.Error(w, fmt.Sprintf(f, a...), http.StatusInternalServerError)
}

func (d *dischargeHandler) badRequest(w http.ResponseWriter, f string, a ...interface{}) {
	http.Error(w, fmt.Sprintf(f, a...), http.StatusBadRequest)
}

type thirdPartyCaveatIdRecord struct {
	RootKey []byte
	Condition string
}

type createResponse struct {
	CaveatId string
}

func (d *dischargeHandler) serveCreate(w http.ResponseWriter, req *http.Request) {
	req.ParseForm()
	condition := req.Form.Get("condition")
	rootKeyStr := req.Form.Get("root-key")

	if len(condition) == 0 || len(rootKeyStr) == 0 {
		d.badRequest(w, "empty values for condition or root key")
		return
	}
	rootKey, err := base64.StdEncoding.DecodeString(rootKeyStr)
	if err != nil {
		d.badRequest(w, "cannot base64-decode root key: %v", err)
		return
	}
	// TODO(rog) what about expiry times?
	idBytes, err := randomBytes(24)
	if err != nil {
		d.internalError(w, "cannot generate random key: %v", err)
		return
	}
	internalId := fmt.Sprintf("third-party-%x", idBytes)
	recordBytes, err := json.Marshal(thirdPartyCaveatIdRecord{
		Condition: condition,
		RootKey: rootKey,
	})
	if err != nil {
		d.internalError(w, "cannot marshal caveat id record: %v", err)
		return
	}
	err = d.store.Put(internalId, string(recordBytes))
	if err != nil {
		d.internalError(w, "cannot store caveat id record: %v", err)
		return
	}
	tpidBytes, err := json.Marshal(&ThirdPartyCaveatId{
		Id: internalId,
	})
	if err != nil {
		d.internalError(w, "cannot marshal caveat id: %v", err)
		return
	}
	respBytes, err := json.Marshal(createResponse{
		CaveatId: base64.StdEncoding.EncodeToString(tpidBytes),
	})
	if err != nil {
		d.internalError(w, "cannot marshal caveat response: %v", err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(respBytes)	
}

func (d *dischargeHandler) servePublicKey(w http.ResponseWriter, r *http.Request) {
	// TODO(rog) implement this
	http.Error(w, "not implemented yet", http.StatusNotImplemented)
}

func randomBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		return nil, fmt.Errorf("cannot generate %d random bytes: %v", n, err)
	}
	return b, nil
}