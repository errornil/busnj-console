package main

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"

	"github.com/chuhlomin/busnj-console/applemaps"
	njt "github.com/chuhlomin/njtransit"
)

var (
	frontendURL url.URL
	proxy       *httputil.ReverseProxy
	busData     *njt.BusDataClient
	maps        *applemaps.TokenGenerator
	// receiptsPath string
)

func check(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
		panic(fmt.Sprintf("%s: %s", msg, err))
	}
}

func handleError(w http.ResponseWriter, r *http.Request, err error, code int, message string) {
	if err != nil {
		message = fmt.Sprintf("%s: %s", message, err)
	}

	log.Printf("%s %s %s %d %s: %v", r.RemoteAddr, r.Method, r.URL, code, message, err)

	errorBytes := []byte("test")

	w.WriteHeader(code)
	fmt.Fprintf(w, "%s", errorBytes)
}

func logMiddleware(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s %s", r.RemoteAddr, r.Method, r.URL)
		handler.ServeHTTP(w, r)
	})
}

func handlerRroxy(w http.ResponseWriter, r *http.Request) {
	req := r
	// Update the headers to allow for SSL redirection
	req.URL.Host = frontendURL.Host
	req.URL.Scheme = frontendURL.Scheme
	req.Header.Set("X-Forwarded-Host", req.Header.Get("Host"))
	req.Host = frontendURL.Host

	// Note that ServeHttp is non blocking and uses a go routine under the hood
	proxy.ServeHTTP(w, req)
}

func main() {
	log.Println("Starting...")

	port := os.Getenv("PORT")

	frontendURL, err := url.Parse(os.Getenv("FRONTEND_ADDRESS"))
	check(err, "Failed to parse URL")

	proxy = httputil.NewSingleHostReverseProxy(frontendURL)

	busData = njt.NewBusDataClient(
		os.Getenv("BUSDATA_USERNAME"),
		os.Getenv("BUSDATA_PASSWORD"),
		njt.BusDataProdURL,
	)

	maps, err = applemaps.NewTokenGenerator(
		os.Getenv("MAPS_PATH_TO_KEY"),
		os.Getenv("MAPS_KEY_ID"),
		os.Getenv("MAPS_TEAM_ID"),
		os.Getenv("MAPS_ORIGIN"),
	)
	check(err, "Failed to create Apple Maps client")

	// todo: refactor
	// receiptsPath = os.Getenv("PATH_TO_RECEIPTS")
	// log.Println("Using receipts path:", receiptsPath)
	// http.Handle("/receipts/src/", http.StripPrefix("/receipts/src/", http.FileServer(http.Dir(receiptsPath))))
	// http.HandleFunc("/receipts/history/", handlerHistory)
	// http.HandleFunc("/receipts/transactions/", handlerTransactions)

	hub := newHub()
	go hub.run()

	http.HandleFunc("/busVehicleDataStream", func(w http.ResponseWriter, r *http.Request) {
		busVehicleDataStream(hub, w, r)
	})
	http.HandleFunc("/gettoken", getToken)
	http.HandleFunc("/", handlerRroxy)

	log.Printf("Listening on port %s...", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), logMiddleware(http.DefaultServeMux)))
}
