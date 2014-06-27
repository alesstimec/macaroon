package main

const AuthServerLocation = "http://localhost:8080/macaroon"

func main() {
	svc := bakery.NewService(
		"http://localhost:8080/",
		bakery.NewMemStorage(),
		checkers.Std,
		httpbakery.NewCaveatIdMaker(nil),
	)

	h := httpbakery.NewHandler(http.HandlerFunc(myHandler), newChecker)
	http.Handle("/")

	// in authorization service:
	svc.AddDischargeHandler(http.DefaultServeMux, thirdPartyChecker, "/macaroon")

	log.Fatal(http.ListenAndServe("localhost:8080", nil))
}

func myHandler(req0 *http.Request, w http.ResponseWriter) {
	req := httpbakery.GetRequest(req0)
	if err := req.Check(canAccessMe); err != nil {
		req.WriteError(w, err)
		return
	}
	fmt.Fprintf(w, "success\n")
}

var canAccessMe = &httpbakery.Capability{
	Id: "can-access-me",
	Caveats: []bakery.Caveat{
		// TODO this won't work - perhaps we
		// should have a function that creates the
		// caveats when needed instead of a literal slice.
		checkers.ExpiresBefore(time.Now().Add(5*time.Minute)),
		checkers.ThirdParty(AuthServerLocation, "access-allowed"),
	},
}

func newChecker(req *bakery.Request) bakery.FirstPartyCaveatChecker {
	return checkers.Map{
		"peer-is": func(s string) error {
			_, arg, _ := checkers.ParseCondition(s)
			if httpHostMatches(req.RemoteAddr, arg) {
				return nil
			}
			return fmt.Errorf("host not allowed")
		},
	}
}

// ---------------------------
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
