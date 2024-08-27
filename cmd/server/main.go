package main

import (
	"log"
	"os"
	"os/signal"
	"flag"
    amqp "github.com/rabbitmq/amqp091-go"
	"github.com/roman-hushpit/learn-pub-sub-starter/internal/pubsub"
	"github.com/roman-hushpit/learn-pub-sub-starter/internal/routing"
	"github.com/roman-hushpit/learn-pub-sub-starter/internal/gamelogic"
)


var (
	uri          = flag.String("uri", "amqp://guest:guest@localhost:5672/", "AMQP URI")
	ErrLog       = log.New(os.Stderr, "[ERROR] ", log.LstdFlags|log.Lmsgprefix)
	Log          = log.New(os.Stdout, "[INFO] ", log.LstdFlags|log.Lmsgprefix)
)

func main() {
	Log.Println("Starting Peril server...")

	conn, err := amqp.Dial(*uri)
	if err != nil {
		ErrLog.Fatalf("producer: error in dial: %s", err)
	}
	defer conn.Close()
	
	Log.Println("Connection was successful")

	channel, _ := conn.Channel()

	Log.Println("Message was sent")
	gamelogic.PrintServerHelp()

	_, queue, err := pubsub.DeclareAndBind(
		conn,
		routing.ExchangePerilTopic,
		routing.GameLogSlug,
		routing.GameLogSlug,
		pubsub.Durable,
	)

	Log.Println(queue.Name)

	for ;; {
		words := gamelogic.GetInput()
		if len(words) == 0 {
			continue
		}
		if words[0] == "pause" {
			Log.Println("you're sending a pause message")
			pubsub.PublishJSON(channel, routing.ExchangePerilDirect, 
				routing.PauseKey, 
				routing.PlayingState {
				IsPaused: true,
			})
			continue
		}
		if words[0] == "resume" {
			Log.Println("you're sending a resume message")
			pubsub.PublishJSON(channel, routing.ExchangePerilDirect, 
				routing.PauseKey, 
				routing.PlayingState {
				IsPaused: false,
			})
			continue
		}
		if words[0] == "quit" {
			Log.Println("Exiting")
			break
		}
		Log.Println("Don't understand the command")
	}
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	<-signalChan
	Log.Println("Program is shutting down")
}
