package main

import (
	"log"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	const rbmqConnString = "amqp://guest:guest@localhost:5672/"

	// open connection to rabbitmq
	conn, err := amqp.Dial(rbmqConnString)
	if err != nil {
		log.Fatalf("Couldn't establish connection with RabbitMQ: %v", err)
	}
	defer conn.Close()

	ch, _, err := pubsub.DeclareAndBind(conn,
		routing.ExchangePerilTopic,
		routing.GameLogSlug,
		"game_logs.*",
		pubsub.SimpleQueueDurable)
	if err != nil {
		log.Fatalf("couldn't declare and bind new queue: %v", err)
	}

	gamelogic.PrintServerHelp()
	for {
		input := gamelogic.GetInput()
		if len(input) == 0 {
			continue
		}

		switch input[0] {
		case "pause":
			log.Println("Publishing pause message...")
			err = pubsub.PublishJSON(ch, routing.ExchangePerilDirect, routing.PauseKey, routing.PlayingState{IsPaused: true})
			if err != nil {
				log.Printf("couldn't publish pause message: %v", err)
				continue
			}
			log.Println("Pause message sent successfully")
		case "resume":
			log.Println("Publishing resume message...")
			err = pubsub.PublishJSON(ch, routing.ExchangePerilDirect, string(routing.PauseKey), routing.PlayingState{IsPaused: false})
			if err != nil {
				log.Printf("couldn't publish resume message: %v", err)
				continue
			}
			log.Println("Resume message sent successfully")
		case "quit":
			log.Println("Exiting... Bye!")
			return
		default:
			log.Println("No command found")
		}
	}
}
