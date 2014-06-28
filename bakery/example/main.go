package main

func main() {
	authEndpoint := serve(authServer)
	serverEndpoint := serve(func() (http.Handler, error) {
		return targetService(authEndpoint)
	})
	client(serverEndpoint)
}

func serve(newHandler func() (http.Handler, error)) (endpointURL string) {
	handler, err := newHandler()
	if err != nil {
		log.Fatal(err)
	}
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		log.Fatal(err)
	}
	go http.Serve(listener, handler)
	return "http://" + listener.Addr().String()
}
