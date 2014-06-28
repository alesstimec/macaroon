package main

// Authorization service.
// This service can act as a checker for third party caveats.

func authService(endpoint string) http.Handler {
	mux := http.NewServeMux()
	svc.AddDischargeHandler(mux, thirdPartyChecker, "/")
	return mux
}

func thirdPartyChecker(req *http.Request, condition string) ([]bakery.Caveat, error) {
	if condition != "access-allowed" {
		return bakery.ErrCaveatNotRecognized
	}
	// TODO check that the HTTP request has cookies that prove
	// something about the client.
	return []bakery.Caveat{{
		Condition: "peer-is localhost",
	}}, nil
}
