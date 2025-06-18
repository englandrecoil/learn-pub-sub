package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	fmt.Println("Starting Peril server...")
	const rbmqConnString = "amqp://guest:guest@localhost:5672/"

	// open connection to rabbitmq
	conn, err := amqp.Dial(rbmqConnString)
	if err != nil {
		log.Fatalf("Couldn't establish connection with RabbitMQ: %v", err)
	}
	defer conn.Close()
	fmt.Println("Server successfully connected to RabbitMQ!")

	pubCh, err := conn.Channel()
	if err != nil {
		log.Fatalf("Couldn't create channel: %v", err)
	}

	err = pubsub.PublishJSON(pubCh, routing.ExchangePerilDirect, routing.PauseKey, routing.PlayingState{
		IsPaused: true,
	})
	if err != nil {
		log.Printf("Сouldn't publish time: %v", err)
	}

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	<-signalChan
	fmt.Println("\nClosing connection...")
}
