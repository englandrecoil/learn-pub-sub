package main

import (
	"fmt"
	"log"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	const rbmqConnString = "amqp://guest:guest@localhost:5672/"

	conn, err := amqp.Dial(rbmqConnString)
	if err != nil {
		log.Fatalf("Couldn't open connection with RabbitMQ: %v", err)
	}
	defer conn.Close()

	username, err := gamelogic.ClientWelcome()
	if err != nil {
		log.Fatalf("couldn't get username: %v", err)
	}
	_, _, err = pubsub.DeclareAndBind(conn,
		routing.ExchangePerilDirect,
		routing.PauseKey+"."+username,
		routing.PauseKey,
		pubsub.SimpleQueueTransient)
	if err != nil {
		log.Println(err)
	}

	state := gamelogic.NewGameState(username)
	gamelogic.PrintClientHelp()
	for {
		input := gamelogic.GetInput()
		switch input[0] {
		case "spawn":
			state.CommandSpawn(input)
		case "move":
			_, err := state.CommandMove(input)
			if err != nil {
				fmt.Println(err)
				continue
			}
		case "status":
			state.CommandStatus()
		case "help":
			gamelogic.PrintClientHelp()
		case "spam":
			log.Println("Spamming not allowed yet!")
		case "quit":
			gamelogic.PrintQuit()
			return
		default:
			log.Printf("command not found")
		}
	}

}
