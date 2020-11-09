package main

import (
	"log"
	"net"
	"net/http"
	"os"

	"github.com/wool/go2hw11/cmd/server_new/app"
	"github.com/wool/go2hw11/pkg/card"
)

const defaultPort = "9999"
const defaultHost = "0.0.0.0"

func main() {
	port, ok := os.LookupEnv("PORT")
	if !ok {
		port = defaultPort
	}

	host, ok := os.LookupEnv("HOST")
	if !ok {
		host = defaultHost
	}

	log.Println(host)
	log.Println(port)

	if err := execute(net.JoinHostPort(host, port)); err != nil {
		log.Println(err)
		os.Exit(1)
	}
}

func execute(addr string) (err error) {
	cardSvc := card.NewService()
	cardSvc.SetCards(card.InitCardsHW11()) // инициализация карт - один раз при запуске приложения

	mux := http.NewServeMux()
	application := app.NewServer(cardSvc, mux)
	application.Init()

	server := &http.Server{
		Addr:    addr,
		Handler: application,
	}
	return server.ListenAndServe()
}
