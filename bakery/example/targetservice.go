package main

type myServer struct {
	svc *bakery.Service
	authEndpoint string
	endpoint
}

func targetService(endpoint, authEndpoint string) (http.Handler, error) {
	srv := &myServer{
		svc: bakery.NewService(bakery.NewServiceParams{
			Location: endpoint,
			NewCaveatIdMaker: httpbakery.NewCaveatIdMaker(nil),
		}),
		authEndpoint: authEndpoint,
	}
	return srv
}

func (srv *myServer) ServeHTTP(req *http.Request, w http.ResponseWriter) {
	breq := srv.svc.NewRequest(srv.checkers(req))
	if err := breq.Check("can-access-me"); err != nil {
		srv.writeError(err)
		return
	}
	fmt.Fprintf(w, "success\n")
}

func (svc *myServer) checkers(req *http.Request) bakery.FirstPartyChecker {
	return checkers.Map{
		"remote-address": func(s string) error {
			// TODO(rog) do we want to distinguish between
			// the two kinds of errors below?
			_, arg, err := checkers.ParseCondition(s)
			if err != nil {
				return err
			}
			if req.RemoteHost != addr {
				return fmt.Errorf("remote address mismatch (need %q)", addr)
			}
		},
	}
}

func (srv *myServer) writeError(w http.ResponseWriter, err error) {
	fail := func(code int, msg string, args ...interface{}) {
		if code == StatusInternalServerError {
			msg = "internal error: " + msg
		}
		http.Error(w, fmt.Sprintf(msg, args...), code)
	}

	verr, _ := err.(*bakery.VerificationError)
	if verr == nil {
		fail(http.StatusForbidden, "%v", err)
		return
	}

	// Work out what caveats we need to apply for the given capability.
	var caveats []bakery.Caveat
	switch verr.RequiredCapability {
	case "can-access-me":
		caveats = []bakery.Caveat{
			checkers.TimeBefore(time.Now().Add(5 * time.Minute)),
			checkers.ThirdParty(srv.authEndpoint, "access-allowed"),
		}, 
	default:
		fail(http.StatusInternalServerError, "capability %q not recognised", verr.RequiredCapability)
		return
	}
	// Mint an appropriate macaroon and send it back to the client.
	m, err := srv.svc.NewMacaroon(verr.RequiredCapability, caveats)
	if err != nil {
		fail(http.StatusInternalServerError, "cannot mint macaroon: %v", err)
		return
	}
	httpbakery.WriteDischargeRequiredError(w, m)
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
