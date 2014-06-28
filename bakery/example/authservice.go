package main

// defined in authorization service

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
