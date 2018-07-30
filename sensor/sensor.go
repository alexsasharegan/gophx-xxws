// Package sensor abstracts over the MPU6050 sensor over I²C
// Data sheets:
// https://www.invensense.com/products/motion-tracking/6-axis/mpu-6050/
// Main resource:
// http://blog.bitify.co.uk/2013/11/reading-data-from-mpu-6050-on-raspberry.html
package sensor

import (
	"encoding/binary"
	"fmt"
	"math"

	"periph.io/x/periph/conn/i2c"
	"periph.io/x/periph/conn/i2c/i2creg"
	"periph.io/x/periph/conn/mmr"
	"periph.io/x/periph/host"
)

const (
	// coefficient for converting radians to degrees
	radToDeg = 180 / math.Pi

	gBase    = 1024
	scale2g  = gBase * 16
	scale4g  = gBase * 8
	scale8g  = gBase * 4
	scale16g = gBase * 2

	// FS_SEL  Full Scale Range  LSB Sensitivity
	// 0       ± 250  °/s        131  LSB/°/s
	// 1       ± 500  °/s        65.5 LSB/°/s
	// 2       ± 1000 °/s        32.8 LSB/°/s
	// 3       ± 2000 °/s        16.4 LSB/°/s
	lsbSensitivity = 131

	// MPU-6050 Registers
	pwrMgmt1 = 0x6b
	pwrMgmt2 = 0x6c
)

var gyros = [3]uint8{0x43, 0x45, 0x47}
var accel = [3]uint8{0x3b, 0x3d, 0x3f}

// Accelerometer represents a sensor connection.
type Accelerometer struct {
	bus  i2c.BusCloser
	conn *i2c.Dev
	mmr  *mmr.Dev8
}

// Open initializes the sensor and connects.
func (a *Accelerometer) Open() error {
	if _, err := host.Init(); err != nil {
		return err
	}

	bus, err := i2creg.Open("1")
	if err != nil {
		return err
	}

	conn := &i2c.Dev{Addr: 0x68, Bus: bus}
	a.bus = bus
	a.conn = conn
	a.mmr = &mmr.Dev8{Conn: conn, Order: binary.BigEndian}

	return a.wake()
}

func (a *Accelerometer) wake() error {
	return a.conn.Tx([]byte{0x6b}, nil)
}

// Close closes the i2c bus.
func (a *Accelerometer) Close() error {
	if err := a.bus.Close(); err != nil {
		return err
	}

	return nil
}

func (a *Accelerometer) readAccel() ([]float64, error) {
	data := make([]float64, len(accel))
	label := []string{"x", "y", "z"}
	fmt.Println("Reading raw acceleration values...")

	for i, reg := range accel {
		v, err := a.mmr.ReadUint16(reg)
		if err != nil {
			return nil, err
		}

		fmt.Println(
			"raw ", label[i], ": ", v,
			" 2c: ", float64From2C(v),
			" scaled: ", float64From2C(v)/scale2g,
		)
		data[i] = float64From2C(v) / scale2g
	}

	return data, nil
}

func (a *Accelerometer) readGyro() ([]float64, error) {
	data := make([]float64, len(gyros))
	label := []string{"x", "y", "z"}
	fmt.Println("Reading raw gyro values...")

	for i, reg := range gyros {
		v, err := a.mmr.ReadUint16(reg)
		if err != nil {
			return nil, err
		}

		fmt.Println(
			"raw ", label[i], ": ", v,
			" 2c: ", float64From2C(v),
			" scaled: ", float64From2C(v)/lsbSensitivity,
		)
		data[i] = float64From2C(v) / lsbSensitivity
	}

	return data, nil
}

// GetGyro reads the current gyroscope data from the sensor,
// and then returns a struct holding the parsed values.
func (a *Accelerometer) GetGyro() (Gyro, error) {
	var gyro Gyro
	d, err := a.readGyro()
	if err != nil {
		return gyro, err
	}

	gyro.data = d

	return gyro, nil
}

// GetAcceleration reads the current acceleration data from the sensor,
// and then returns a struct holding the parsed values.
func (a *Accelerometer) GetAcceleration() (Acceleration, error) {
	var acc Acceleration
	d, err := a.readAccel()
	if err != nil {
		return acc, err
	}

	acc.data = d

	return acc, nil
}

// Gyro represents a single readout of gyroscope data.
type Gyro struct {
	data []float64
}

// GetValues returns the raw x, y, z values parsed from the sensor.
func (acc Gyro) GetValues() (x, y, z float64) {
	return acc.data[0], acc.data[1], acc.data[2]
}

// Acceleration represents a single readout of acceleration data.
type Acceleration struct {
	data []float64
}

// GetValues returns the raw x, y, z values parsed from the sensor.
func (acc Acceleration) GetValues() (x, y, z float64) {
	return acc.data[0], acc.data[1], acc.data[2]
}

// GetXRotation returns the degree rotation
func (acc Acceleration) GetXRotation() float64 {
	x, y, z := acc.GetValues()
	rad := math.Atan2(y, distance(x, z))
	fmt.Println("math.Atan2(y, distance(x, z)): ", rad)
	// radians -> degrees
	return rad * radToDeg
}

// GetYRotation returns the degree rotation
func (acc Acceleration) GetYRotation() float64 {
	x, y, z := acc.GetValues()
	rad := math.Atan2(x, distance(y, z))
	fmt.Println("math.Atan2(x, distance(y, z))", rad)
	// radians -> degrees
	return -(rad * radToDeg)
}

func distance(a, b float64) float64 {
	return math.Sqrt((a * a) + (b * b))
}

func float64From2C(x uint16) float64 {
	if x>>15 == 1 {
		return -float64(x ^ 0xFFFF + 1)
	}

	return float64(x)
}

func intFrom2C(x uint16) int {
	if x>>15 == 1 {
		return -int(x ^ 0xFFFF + 1)
	}

	return int(x)
}
