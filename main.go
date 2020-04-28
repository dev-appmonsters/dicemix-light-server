package main

import (
	"flag"
	"net/http"

	"github.com/dev-appmonsters/dicemix-light-server/server"

	log "github.com/sirupsen/logrus"
)

var addr = flag.String("addr", ":8082", "http service address")

func main() {
	// setup logger
	formatter := &log.TextFormatter{
		FullTimestamp: true,
	}
	log.SetFormatter(formatter)

	flag.Parse()

	connection := server.NewConnection()

	log.Info("Server Started")
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		connection.Register(w, r)
	})

	if err := http.ListenAndServe(*addr, nil); err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
