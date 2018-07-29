package sensor

import (
	"math/rand"
)

// State represents the current state of an accelerometer.
type State struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
	Z float64 `json:"z"`
}

const x = 100

// RandData returns dummy SensorData.
func RandData() State {
	return State{rand.Float64() * x, rand.Float64() * x, rand.Float64() * x}
}
