package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/chuhlomin/busnj-console/pkg/proxy"
	"github.com/chuhlomin/busnj-console/pkg/redis"
	"github.com/chuhlomin/busnj-console/pkg/websocket"

	"github.com/caarlos0/env"
)

type config struct {
	Port            string `env:"PORT" envDefault:"6001"`
	FrontendAddress string `env:"FRONTEND_ADDRESS" envDefault:"http://busnj-console-ui/"`
	RedisNetwork    string `env:"REDIS_NETWORK" envDefault:"tcp"`
	RedisAddr       string `env:"REDIS_ADDR" envDefault:"redis:6379"`
	RedisSize       int    `env:"REDIS_SIZE" envDefault:"10"`
}

func check(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
		panic(fmt.Sprintf("%s: %s", msg, err))
	}
}

func logMiddleware(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s %s", r.RemoteAddr, r.Method, r.URL)
		handler.ServeHTTP(w, r)
	})
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

func handlerRroxy(proxy *proxy.Client, w http.ResponseWriter, r *http.Request) {
	proxy.Serve(w, r)
}

// handlerBusVehicleData returns list of known bus vehicles
func handlerBusVehicleData(redis *redis.Client, w http.ResponseWriter, r *http.Request) {
	data, err := redis.LoadBusVehicleDataMessages()
	if err != nil {
		log.Printf("Failed to load BusVehicleDataMessages: %v", err)
		return
	}

	response, err := json.Marshal(data)
	if err != nil {
		log.Printf("Failed to marshal BusVehicleData result: %v", err)
		return
	}
	w.Write(response)
}

// handlerBusVehicleDataStream handles websocket requests from the peer.
func handlerBusVehicleDataStream(hub *websocket.Hub, w http.ResponseWriter, r *http.Request) {
	client, err := websocket.NewClient(hub, w, r)
	if err != nil {
		log.Printf("Failed to create WebSocket client: %v", err)
		return
	}

	hub.Register <- client

	go client.WritePump()
	go client.ReadPump()
}

func main() {
	log.Println("Starting...")

	c := config{}
	err := env.Parse(&c)

	port := c.Port

	proxy, err := proxy.NewClient(c.FrontendAddress)
	check(err, "Failed to create Proxy client")

	redis, err := redis.NewClient(
		c.RedisNetwork,
		c.RedisAddr,
		c.RedisSize,
	)
	check(err, "Failed to create Redis client")

	hub := websocket.NewHub()
	go hub.Run()

	go redis.ConsumeBusVehicleDataChannel(hub)

	http.HandleFunc("/busVehicleDataStream", func(w http.ResponseWriter, r *http.Request) {
		handlerBusVehicleDataStream(hub, w, r)
	})
	http.HandleFunc("/busVehicleData", func(w http.ResponseWriter, r *http.Request) {
		handlerBusVehicleData(redis, w, r)
	})
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		handlerRroxy(proxy, w, r)
	})

	log.Printf("Listening on port %s...", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), logMiddleware(http.DefaultServeMux)))
}
