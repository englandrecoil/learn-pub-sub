package main

import (
	"fmt"
	"time"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)

func handlerPause(gs *gamelogic.GameState) func(routing.PlayingState) pubsub.AckType {
	return func(state routing.PlayingState) pubsub.AckType {
		defer fmt.Print("> ")
		gs.HandlePause(state)
		return pubsub.Ack
	}
}

func handlerMove(gs *gamelogic.GameState, publishChannel *amqp.Channel) func(am gamelogic.ArmyMove) pubsub.AckType {
	return func(am gamelogic.ArmyMove) pubsub.AckType {
		defer fmt.Print("> ")
		outcomeType := gs.HandleMove(am)
		switch outcomeType {
		case gamelogic.MoveOutComeSafe:
			return pubsub.Ack
		case gamelogic.MoveOutcomeMakeWar:
			err := pubsub.PublishJSON(
				publishChannel,
				string(routing.ExchangePerilTopic),
				string(routing.WarRecognitionsPrefix)+"."+gs.GetUsername(),
				gamelogic.RecognitionOfWar{
					Attacker: am.Player,
					Defender: gs.GetPlayerSnap(),
				},
			)
			if err != nil {
				fmt.Printf("error: %v\n", err)
				return pubsub.NackRequeue
			}
			return pubsub.Ack
		case gamelogic.MoveOutcomeSamePlayer:
			return pubsub.NackDiscard
		}
		fmt.Println("error: unknown move outcome")
		return pubsub.NackDiscard
	}
}

func handleWar(gs *gamelogic.GameState, publishCh *amqp.Channel) func(rw gamelogic.RecognitionOfWar) pubsub.AckType {
	return func(rw gamelogic.RecognitionOfWar) pubsub.AckType {
		outcome, winner, loser := gs.HandleWar(rw)
		switch outcome {
		case gamelogic.WarOutcomeNotInvolved:
			return pubsub.NackRequeue
		case gamelogic.WarOutcomeNoUnits:
			return pubsub.NackDiscard
		case gamelogic.WarOutcomeOpponentWon:
			err := pubsub.PublishGob(
				publishCh,
				string(routing.ExchangePerilTopic),
				string(routing.GameLogSlug)+"."+gs.GetUsername(),
				routing.GameLog{
					CurrentTime: time.Now(),
					Message:     fmt.Sprintf("%s won a war against %s", winner, loser),
					Username:    gs.GetUsername(),
				},
			)
			if err != nil {
				return pubsub.NackRequeue
			}
			return pubsub.Ack
		case gamelogic.WarOutcomeYouWon:
			err := pubsub.PublishGob(
				publishCh,
				string(routing.ExchangePerilTopic),
				string(routing.GameLogSlug)+"."+gs.GetUsername(),
				routing.GameLog{
					CurrentTime: time.Now(),
					Message:     fmt.Sprintf("%s won a war against %s", winner, loser),
					Username:    gs.GetUsername(),
				},
			)
			if err != nil {
				return pubsub.NackRequeue
			}
			return pubsub.Ack
		case gamelogic.WarOutcomeDraw:
			err := pubsub.PublishGob(
				publishCh,
				string(routing.ExchangePerilTopic),
				string(routing.GameLogSlug)+"."+gs.GetUsername(),
				routing.GameLog{
					CurrentTime: time.Now(),
					Message:     fmt.Sprintf("A war between %s and %s resulted in a draw", winner, loser),
					Username:    gs.GetUsername(),
				})
			if err != nil {
				return pubsub.NackRequeue
			}
			return pubsub.Ack
		}
		fmt.Println("error: unknown war outcome")
		return pubsub.NackDiscard
	}
}
