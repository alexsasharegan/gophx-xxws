package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/alexsasharegan/gophx-xxws/sensor"
	"github.com/gobuffalo/packr"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/alexsasharegan/gophx-xxws/ws"
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

	ticker := time.NewTicker(time.Millisecond * 16)

MainLoop:
	for {
		select {
		case <-ticker.C:
			b, err := json.Marshal(sensor.RandData())
			if err != nil {
				log.Println(fmt.Sprintf("Error serializing json: %v", err))
				break
			}
			hub.Broadcast(b)
		case <-sig:
			ticker.Stop()
			log.Println("Closing ws connections...")
			hub.Close()

			break MainLoop
		}
	}

	log.Println("Goodbye 👋")
}
