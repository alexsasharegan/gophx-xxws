package sensor

import (
	"math/rand"
)

// State represents the current state of an accelerometer.
type State struct {
	x float64
	y float64
	z float64
}

const x = 100

// RandData returns dummy SensorData.
func RandData() State {
	return State{rand.Float64() * x, rand.Float64() * x, rand.Float64() * x}
}
