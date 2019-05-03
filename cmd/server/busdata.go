package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	njt "github.com/chuhlomin/njtransit"
)

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

func busVehicleDataStream(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	defer c.Close()

	mt, message, err := c.ReadMessage()
	if err != nil {
		log.Println("read error:", err)
		return
	}
	log.Printf("← %s", message)

	rc := make(chan njt.BusVehicleDataRow)
	ec := make(chan error)
	go busData.GetBusVehicleDataStream(rc, ec, 5*time.Second, true)

	for {
		select {
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

			response, err := json.Marshal(message)
			if err != nil {
				log.Printf("Failed to marshal BusVehicleDataMessage: %v", err)
				return
			}

			err = c.WriteMessage(mt, response)
			if err != nil {
				log.Printf("Failed to write message: %v", err)
				return
			}
			log.Printf("→ %s", string(response))

		case err := <-ec: // errors in the library
			fmt.Println(err)
		}
	}
}
