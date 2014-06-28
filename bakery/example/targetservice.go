package main

type myServer struct {
	svc *bakery.Service
}

func targetService(authEndpoint string) (http.Handler, error) {
	svc := bakery.NewService(
		"http://localhost:8080/",
		bakery.NewMemStorage(),
		checkers.Std,
		httpbakery.NewCaveatIdMaker(nil),
	)
	srv := &myServer{
		svc: svc,
	}

	h := httpbakery.NewHandler(srv, newChecker)
	http.Handle("/")

	// in authorization service:
	svc.AddDischargeHandler(http.DefaultServeMux, thirdPartyChecker, "/macaroon")

	log.Fatal(http.ListenAndServe("localhost:8080", nil))
}

func (svc *myServer) checkers(req *http.Request) bakery.FirstPartyChecker {
	return checkers.Map{
		"remote-address": checkers.Magic(func(_, addr string) error {
			if req.RemoteHost != addr {
				return fmt.Errorf("remote address does not match %q", addr)
			}
		},
	}
}

func (srv *myServer) ServeHTTP(req *http.Request, w http.ResponseWriter) {
	breq := srv.svc.NewRequest(srv.checkers(req))
	if err := breq.Check("can-access-me"); err != nil {
		srv.writeError(err)
		return
	}
	fmt.Fprintf(w, "success\n")
}

func (srv *myServer) writeError(w http.ResponseWriter, err error) {
	verr, _ := err.(*bakery.VerificationError)
	if verr == nil {
		http.Error(w, err.Error(), http.CodeFail)
		return
	}
	srv.svc.NewMacaroon("

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
