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
	"github.com/rakyll/statik/fs"

	_ "github.com/alexsasharegan/gophx-xxws/statik"
)

var (
	addr = flag.String("addr", "0.0.0.0:3000", "http service address")
)

func main() {
	flag.Parse()
	statikFS, err := fs.New()
	if err != nil {
		log.Fatalln(err)
	}

	var a sensor.Accelerometer
	if err := a.Open(); err != nil {
		log.Fatalln(err)
	}
	defer a.Close()

	hub := ws.NewHub()
	go hub.RunLoop()

	sig := make(chan os.Signal)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

	// h := http.FileServer(assets)
	// http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
	// 	fmt.Println("received request: ", r.URL.Path)
	// 	h.ServeHTTP(w, r)
	// })
	http.Handle("/", http.FileServer(statikFS))
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		log.Println("[ws] Client connection received.")

		err := ws.ServeWS(hub, w, r)
		if err != nil {
			log.Println(
				fmt.Sprintf("Error upgrading request to ws: %v", err),
			)
		}
	})

	log.Println("Opening server on port ", *addr)
	go func() {
		if err := http.ListenAndServe(*addr, nil); err != nil {
			log.Fatalln(
				fmt.Sprintf("Could not bind server to address '%s'", *addr),
				err,
			)
		}
	}()

	// Blocking forever loop only broken by interrupt/terminate signal.
	broadcastLoop(hub, &a, sig)
	log.Println("Goodbye 👋")
}

func broadcastLoop(hub *ws.Hub, a *sensor.Accelerometer, sig <-chan os.Signal) {
	// Send new data twice per render cycle (60Hz)
	ticker := time.NewTicker(time.Second / 120)
	defer func() {
		ticker.Stop()
		hub.Close()
	}()

	for {
		select {
		case <-ticker.C:
			d, err := getData(a)
			if err != nil {
				log.Println("Error reading sensor data: ", err)
				break
			}
			b, err := json.Marshal(d)
			if err != nil {
				log.Println("Error serializing json: ", err)
				break
			}
			hub.Broadcast(b)
		case s := <-sig:
			log.Println("Received shutdown signal: ", s.String())
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
