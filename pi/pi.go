package main

import (
	"fmt"
	"log"

	"github.com/alexsasharegan/gophx-xxws/sensor"
)

func main() {
	var a sensor.Accelerometer

	if err := a.Open(); err != nil {
		log.Fatalln(err)
	}

	defer a.Close()

	fmt.Println("------")
	fmt.Println("Golang")
	fmt.Println("------")
	fmt.Println()
	fmt.Println("gyro data")
	fmt.Println("---------")

	_, err := a.GetGyro()
	if err != nil {
		log.Fatalln(err)
	}

	fmt.Println()
	fmt.Println("accelerometer data")
	fmt.Println("------------------")

	accel, err := a.GetAcceleration()
	if err != nil {
		log.Fatalln(err)
	}

	fmt.Println("x rotation: ", accel.GetXRotation())
	fmt.Println("y rotation: ", accel.GetYRotation())
}
