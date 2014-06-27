package httpbakery

// DischargeHandler returns an HTTP handler that issues discharge macaroons
// to clients after using the given check function to ensure that the given
// client is allowed to obtain a discharge.
//
// The check function is used to check whether a client making the given
// request should be allowed a discharge for the given caveat. If it
// does not return an error, the caveat will be discharged, with any
// returned caveats also added to the discharge macaroon.
func DischargeHandler(store bakery.Storage, checker ThirdPartyChecker) http.Handler {
	// POST /discharge
	//	params:
	//		id of macaroon to discharge
	//		location of original macaroon?
	//	result:
	//		macaroon in json format
	//		or error
	//
	// POST /create
	//	params:
	//		caveat to discharge
	//	result:
	//		caveat id
	//
	// GET /publickey
	//	result:
	//		public key of service
	//		expiry time of key
}
