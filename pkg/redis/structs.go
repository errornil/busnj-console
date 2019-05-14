package redis

// BusVehicleDataMessage represents information about single Vehicle
type BusVehicleDataMessage struct {
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
