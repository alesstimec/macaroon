package main

func main() {
	svc := bakery.NewService(bakery.NewMemStorage())

	checker := func(req *http.Request, caveat string) {
		something
	}

	//bakery.Checkers(map[string] func(string)error {
	//	"std": checkers.Standard,
	//	"", myChecker,
	//})
	h := httpbakery.NewHandler(http.HandlerFunc(myHandler), checker)
	http.Handle("/")

	// optional
	http.Handle("/macaroon", http.StripPrefix("/macaroon", svc.DischargeHandler))
}

func myHandler(req0 *http.Request, w http.ResponseWriter) {
	req := httpbakery.GetRequest(req0)
	if err := req.Check(
		checkers.ExpiresBefore(time.Now().Add(5*time.Minute)),
		checkers.ThirdParty(AuthServerLocation, "access-allowed"),
	); err != nil {
		req.WriteError(w, err)
		return
	}
}
