package datastruct

import "time"

// Data : Type for all returns from sensorGet functions
type Data struct {
	ID   string
	Time time.Time
	Data string
}
