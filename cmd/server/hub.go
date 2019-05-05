package main

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	njt "github.com/chuhlomin/njtransit"
)

// Hub maintains the set of active clients and broadcasts messages to the
// clients.
type Hub struct {
	// Registered clients.
	clients map[*Client]bool

	// Inbound messages from the clients.
	broadcast chan []byte

	// Register requests from the clients.
	register chan *Client

	// Unregister requests from clients.
	unregister chan *Client

	// Stores all busVehicleDataMessage by VehicleID GetBusVehicleData
	busVehicleDataStore map[string]busVehicleDataMessage
}

type busVehicleDataMessage struct {
	VehicleID            string `json:"vehicleID"`
	Route                string `json:"route"`                // 1
	RunID                string `json:"runID"`                // 21
	TripBlock            string `json:"tripBlock"`            // 001HL064
	PatternID            string `json:"patternID"`            // 264
	Destination          string `json:"destination"`          // 1 NEWARK-IVY HILL VIA RIVER TERM
	Longitude            string `json:"longitude"`            // -74.24513778686523
	Latitude             string `json:"latitude"`             // 40.73779029846192
	GPSTimestmp          string `json:"GPStimestmp"`          // 25-Apr-2019 12:15:12 AM
	LastModified         string `json:"lastModified"`         // 25-Apr-2019 12:16:10 AM
	AsInternalTripNumber string `json:"asInternalTripNumber"` // 13734490
	// Timepoints           []BusVehicleDataRowTimepoint `json:"timepoints"`
}

func newHub() *Hub {
	return &Hub{
		broadcast:           make(chan []byte),
		register:            make(chan *Client),
		unregister:          make(chan *Client),
		clients:             make(map[*Client]bool),
		busVehicleDataStore: map[string]busVehicleDataMessage{},
	}
}

func (h *Hub) run() {
	rc := make(chan njt.BusVehicleDataRow)
	ec := make(chan error)
	go busData.GetBusVehicleDataStream(rc, ec, 5*time.Second, true)

	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
		case row := <-rc:
			message := busVehicleDataMessage{
				VehicleID:            row.VehicleID,
				Route:                row.Route,
				RunID:                row.RunID,
				TripBlock:            row.TripBlock,
				PatternID:            row.PatternID,
				Destination:          strings.TrimSpace(row.Destination),
				Longitude:            row.Longitude,
				Latitude:             row.Latitude,
				GPSTimestmp:          row.GPSTimestmp,
				LastModified:         row.LastModified,
				AsInternalTripNumber: row.AsInternalTripNumber,
			}

			h.busVehicleDataStore[row.VehicleID] = message

			response, err := json.Marshal(message)
			if err != nil {
				log.Printf("Failed to marshal BusVehicleDataMessage: %v", err)
				continue
			}

			for client := range h.clients {
				client.send <- response
			}

		case err := <-ec: // errors in the library
			fmt.Println(err)
		}
	}
}

// GetBusVehicleData returns all known Vehicles at app managed to get since start
func (h *Hub) getBusVehicleData() map[string]busVehicleDataMessage {
	return h.busVehicleDataStore
}
