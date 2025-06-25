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
	publishCh, err := conn.Channel()
	if err != nil {
		log.Fatalf("couldn't open chanel for client: %v", err)
	}

	username, err := gamelogic.ClientWelcome()
	if err != nil {
		log.Fatalf("couldn't get username: %v", err)
	}
	gs := gamelogic.NewGameState(username)

	err = pubsub.SubscribeJSON(conn,
		routing.ExchangePerilDirect,
		routing.PauseKey+"."+gs.GetUsername(),
		routing.PauseKey,
		pubsub.SimpleQueueTransient,
		handlerPause(gs),
	)
	if err != nil {
		log.Fatalf("couldn't subscribe consumer to pause queue: %v", err)
	}

	err = pubsub.SubscribeJSON(conn,
		string(routing.ExchangePerilTopic),
		string(routing.ArmyMovesPrefix)+"."+gs.GetUsername(),
		string(routing.ArmyMovesPrefix)+".*",
		pubsub.SimpleQueueTransient,
		handlerMove(gs, publishCh),
	)
	if err != nil {
		log.Fatalf("couldn't subscribe consumer to army_moves queue: %v", err)
	}

	err = pubsub.SubscribeJSON(
		conn,
		string(routing.ExchangePerilTopic),
		string(routing.WarRecognitionsPrefix),
		string(routing.WarRecognitionsPrefix)+".*",
		pubsub.SimpleQueueDurable,
		handleWar(gs),
	)
	if err != nil {
		log.Fatalf("couldn't subscribe consumer to war queue: %v", err)
	}

	gamelogic.PrintClientHelp()
	for {
		input := gamelogic.GetInput()
		switch input[0] {
		case "spawn":
			gs.CommandSpawn(input)
		case "move":
			mv, err := gs.CommandMove(input)
			if err != nil {
				fmt.Println(err)
				continue
			}
			err = pubsub.PublishJSON(publishCh, routing.ExchangePerilTopic, routing.ArmyMovesPrefix+"."+mv.Player.Username, mv)
			if err != nil {
				log.Printf("couldn't publish move: %v", err)
			} else {
				fmt.Printf("Moved %v units to %s\n", len(mv.Units), mv.ToLocation)
			}
		case "status":
			gs.CommandStatus()
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
