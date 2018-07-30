// Package sensor abstracts over the MPU6050 sensor over I²C
// Data sheets:
// https://www.invensense.com/products/motion-tracking/6-axis/mpu-6050/
// Main resource:
// http://blog.bitify.co.uk/2013/11/reading-data-from-mpu-6050-on-raspberry.html
package sensor

import (
	"encoding/binary"
	"math"

	"periph.io/x/periph/conn/i2c"
	"periph.io/x/periph/conn/i2c/i2creg"
	"periph.io/x/periph/conn/mmr"
	"periph.io/x/periph/host"
)

const (
	// coefficient for converting radians to degrees
	degToRad = math.Pi / 180
	radToDeg = 180 / math.Pi

	gravBase = 1024
	scale2g  = gravBase * 16
	scale4g  = gravBase * 8
	scale8g  = gravBase * 4
	scale16g = gravBase * 2

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

var (
	gyroRegs  = [3]uint8{0x43, 0x45, 0x47}
	accelRegs = [3]uint8{0x3b, 0x3d, 0x3f}
)

// Accelerometer represents a sensor connection.
type Accelerometer struct {
	bus  i2c.BusCloser
	conn *i2c.Dev
	mmr  *mmr.Dev8
}

// Open initializes the sensor and connects.
func (a *Accelerometer) Open() error {
	// Ensure the periph lib has been initialized. Mutliple calls are safe.
	if _, err := host.Init(); err != nil {
		return err
	}

	// Open an SMBus
	bus, err := i2creg.Open("1")
	if err != nil {
		return err
	}

	// Keep a ref since we are responsible for closing this.
	a.bus = bus
	// Conn implements the periph conn interface.
	// Mostly, it just writes our device register as the first byte in a tx.
	a.conn = &i2c.Dev{Addr: 0x68, Bus: a.bus}
	// Abstraction over our conn that helps us read the bytes returned.
	a.mmr = &mmr.Dev8{Conn: a.conn, Order: binary.BigEndian}

	// The sensor starts in sleep mode.
	if err := a.wake(); err != nil {
		return err
	}

	return nil
}

func (a *Accelerometer) wake() error {
	return a.conn.Tx([]byte{pwrMgmt1}, nil)
}

// Close closes the i2c bus.
func (a *Accelerometer) Close() error {
	if err := a.bus.Close(); err != nil {
		return err
	}

	return nil
}

func (a *Accelerometer) readAccel() ([]float64, error) {
	data := make([]float64, len(accelRegs))

	for i, reg := range accelRegs {
		v, err := a.mmr.ReadUint16(reg)
		if err != nil {
			return nil, err
		}

		data[i] = float64From2C(v) / scale2g
	}

	return data, nil
}

func (a *Accelerometer) readGyro() ([]float64, error) {
	data := make([]float64, len(gyroRegs))

	for i, reg := range gyroRegs {
		v, err := a.mmr.ReadUint16(reg)
		if err != nil {
			return nil, err
		}

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
	return rad * radToDeg
}

// GetYRotation returns the degree rotation
func (acc Acceleration) GetYRotation() float64 {
	x, y, z := acc.GetValues()
	rad := math.Atan2(x, distance(y, z))
	return -(rad * radToDeg)
}

func distance(a, b float64) float64 {
	return math.Sqrt((a * a) + (b * b))
}

func float64From2C(x uint16) float64 {
	// Read the MSB for signedness.
	if x>>15 == 1 {
		// Invert bits, add 1, and add negative sign.
		return -float64(x ^ 0xFFFF + 1)
	}

	return float64(x)
}

func intFrom2C(x uint16) int {
	// Read the MSB for signedness.
	if x>>15 == 1 {
		// Invert bits, add 1, and add negative sign.
		return -int(x ^ 0xFFFF + 1)
	}

	return int(x)
}
