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



func main() {
	fmt.Println("Starting Peril client...")
	conn, err := amqp.Dial(*uri)
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

	gs := gamelogic.NewGameState(username)

	for {
		words := gamelogic.GetInput()
		if len(words) == 0 {
			continue
		}
		switch words[0] {
		case "move":
			_, err := gs.CommandMove(words)
			if err != nil {
				fmt.Println(err)
				continue
			}

			// TODO: publish the move
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
