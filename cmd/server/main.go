package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"

	amqp "github.com/rabbitmq/amqp091-go"
)

const rbmqConnString = "amqp://guest:guest@localhost:5672/"

func main() {
	fmt.Println("Starting Peril server...")

	conn, err := amqp.Dial(rbmqConnString)
	if err != nil {
		log.Fatalf("Couldn't establish connection with RabbitMQ: %v", err)
	}
	defer conn.Close()
	fmt.Println("Server successfully connected to RabbitMQ!")

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	<-signalChan
	fmt.Println("\nClosing connection...")
}
