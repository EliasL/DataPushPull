package getdata

import "time"

// GetData : Type for all returns from sensorGet functions
type GetData struct {
	ID   string
	Time time.Time
	Data string
}
