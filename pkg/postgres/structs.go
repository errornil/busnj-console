package postgres

type Trip struct {
	TripID       int
	RouteID      int
	ServiceID    int
	TripHeadsign string
	DirectionID  int
	BlockID      string
	ShapeID      int
}
