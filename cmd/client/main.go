package main

import (
	"log"
	"flag"
	"os"
	"fmt"
    amqp "github.com/rabbitmq/amqp091-go"
	"github.com/roman-hushpit/learn-pub-sub-starter/internal/gamelogic"
	"github.com/roman-hushpit/learn-pub-sub-starter/internal/routing"
	"github.com/roman-hushpit/learn-pub-sub-starter/internal/pubsub"
)


var (
	uri          = flag.String("uri", "amqp://guest:guest@localhost:5672/", "AMQP URI")
	ErrLog       = log.New(os.Stderr, "[ERROR] ", log.LstdFlags|log.Lmsgprefix)
	Log          = log.New(os.Stdout, "[INFO] ", log.LstdFlags|log.Lmsgprefix)
)


func handlerPause(gs *gamelogic.GameState) func(routing.PlayingState) pubsub.AckType {
	defer fmt.Print("> ")
	var pauseClosure = func (routing routing.PlayingState) pubsub.AckType  {
		gs.HandlePause(routing)
		return  pubsub.Ack
	}
	return pauseClosure
}

func handlerMove(gs *gamelogic.GameState) func(gamelogic.ArmyMove) pubsub.AckType {
	defer fmt.Print("> ")
	var moveClosure = func(move gamelogic.ArmyMove) pubsub.AckType {
		result := gs.HandleMove(move)
		if result == gamelogic.MoveOutComeSafe || 
		result == gamelogic.MoveOutcomeMakeWar{
			return pubsub.Ack
		}
		return pubsub.NackDiscard
	  }
	return moveClosure
}



func main() {
	fmt.Println("Starting Peril client...")
	conn, err := amqp.Dial(*uri)
	channel, _ := conn.Channel()
	defer conn.Close()
	if err != nil {
		ErrLog.Fatalf("producer: error in dial: %s", err)
	}
	Log.Println("Connection was successful")
	username, _ := gamelogic.ClientWelcome()
	_, queue, err := pubsub.DeclareAndBind(
		conn,
		routing.ExchangePerilDirect,
		routing.PauseKey+"."+username,
		routing.PauseKey,
		pubsub.Transient,
	)
	if err != nil {
		log.Fatalf("could not subscribe to pause: %v", err)
	}
	fmt.Printf("Queue %v declared and bound!\n", queue.Name)


	_, moveQueue, err1 := pubsub.DeclareAndBind(
		conn,
		routing.ExchangePerilTopic,
		routing.ArmyMovesPrefix + "." + username,
		routing.ArmyMovesPrefix + ".*",
		pubsub.Transient,
	)
	if err1 != nil {
		log.Fatalf("could not subscribe to pause: %v", err)
	}
	fmt.Printf("Queue %v declared and bound!\n", moveQueue.Name)


	gs := gamelogic.NewGameState(username)

	handlePauseFunc := handlerPause(gs)
	handleMoveFunc := handlerMove(gs)

	pubsub.SubscribeJSON(conn, routing.ExchangePerilDirect, routing.PauseKey+"."+username, routing.PauseKey,
 	pubsub.Transient, handlePauseFunc)

	pubsub.SubscribeJSON(conn, routing.ExchangePerilDirect, routing.ArmyMovesPrefix+"."+username, 
	routing.ArmyMovesPrefix + ".*",
 	pubsub.Transient, handleMoveFunc)



	for {
		words := gamelogic.GetInput()
		if len(words) == 0 {
			continue
		}
		switch words[0] {
		case "move":
			armyMove, err := gs.CommandMove(words)
			if err != nil {
				fmt.Println(err)
				continue
			}

			errorMove := pubsub.PublishJSON(channel, routing.ExchangePerilTopic, 
				routing.ArmyMovesPrefix + "." + username, 
				armyMove)
			if errorMove != nil {
				Log.Fatalln(errorMove)	
			} else {
				Log.Printf("Successfully published a move: %v \n", armyMove)
			}
		case "spawn":
			err = gs.CommandSpawn(words)
			if err != nil {
				fmt.Println(err)
				continue
			}
		case "status":
			gs.CommandStatus()
		case "help":
			gamelogic.PrintClientHelp()
		case "spam":
			// TODO: publish n malicious logs
			fmt.Println("Spamming not allowed yet!")
		case "quit":
			gamelogic.PrintQuit()
			return
		default:
			fmt.Println("unknown command")
		}
	}
}
