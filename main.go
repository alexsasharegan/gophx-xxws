package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/alexsasharegan/gophx-xxws/sensor"
	"github.com/alexsasharegan/gophx-xxws/ws"
	"github.com/gobuffalo/packr"
)

var (
	addr   = flag.String("addr", ":3000", "http service address")
	assets packr.Box
)

func init() {
	flag.Parse()
	assets = packr.NewBox("./www/build")
}

func main() {
	var acc *sensor.Accelerometer
	if err := acc.Open(); err != nil {
		log.Fatalln(err)
	}

	hub := ws.NewHub()
	go hub.RunLoop()

	sig := make(chan os.Signal)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

	http.Handle("/", http.FileServer(assets))
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		err := ws.ServeWS(hub, w, r)
		if err != nil {
			log.Println(
				fmt.Sprintf("Error upgrade request to ws: %v", err),
			)
		}
	})

	log.Println(fmt.Sprintf("Opening server on port %s", *addr))
	go func() {
		if err := http.ListenAndServe(*addr, nil); err != nil {
			log.Fatalln(
				fmt.Sprintf("Could not bind server to address '%s'", *addr),
				err,
			)
		}
	}()

	// Blocking forever loop only broken by interrupt/terminate signal.
	broadcastLoop(hub, sig)
	log.Println("Goodbye 👋")
}

func broadcastLoop(hub *ws.Hub, sig <-chan os.Signal) {
	var a sensor.Accelerometer
	if err := a.Open(); err != nil {
		log.Fatalln("Error opening connection to sensor: ", err)
	}
	// Send new data twice per render cycle (60Hz)
	ticker := time.NewTicker(time.Second / 120)

	for {
		select {
		case <-ticker.C:
			d, err := getData(&a)
			if err != nil {
				log.Println("Error reading sensor data: ", err)
				break
			}
			b, err := json.Marshal(d)
			if err != nil {
				log.Println(fmt.Sprintf("Error serializing json: %v", err))
				break
			}
			hub.Broadcast(b)
		case s := <-sig:
			log.Println("Received shutdown signal: ", s.String())
			ticker.Stop()
			hub.Close()
			a.Close()

			return
		}
	}
}

type accelerationData struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
	Z float64 `json:"z"`

	Rotation []float64 `json:"rotation"`
}

type gyroData struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
	Z float64 `json:"z"`
}

type sensorData struct {
	Acceleration *accelerationData `json:"acceleration"`
	Gyro         *gyroData         `json:"gyro"`
}

func getData(a *sensor.Accelerometer) (*sensorData, error) {
	accel, err := a.GetAcceleration()
	if err != nil {
		return nil, err
	}

	gyro, err := a.GetGyro()
	if err != nil {
		return nil, err
	}

	ax, ay, az := accel.GetValues()
	xr := accel.GetXRotation()
	yr := accel.GetYRotation()
	gx, gy, gz := gyro.GetValues()

	return &sensorData{
		Acceleration: &accelerationData{
			X:        ax,
			Y:        ay,
			Z:        az,
			Rotation: []float64{xr, yr},
		},
		Gyro: &gyroData{
			X: gx,
			Y: gy,
			Z: gz,
		},
	}, nil

}
