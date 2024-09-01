package pubsub

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	//"sync"
	amqp "github.com/rabbitmq/amqp091-go"
)

type QueueType int

const (
	Durable QueueType = iota
	Transient
)

//, wg *sync.WaitGroup
func StartConsumer[T any](deliveries <-chan amqp.Delivery, handler func(T)) {
	//defer wg.Done()
	for delivery := range deliveries {
		var msg T
		// Unmarshal the message body into the generic type T
		err := json.Unmarshal(delivery.Body, &msg)
		if err != nil {
			log.Printf("Error unmarshalling message: %v", err)
			continue
		}

		// Call the handler function with the unmarshalled message
		handler(msg)

		// Acknowledge the message
		err = delivery.Ack(false)
		if err != nil {
			log.Printf("Error acknowledging message: %v", err)
		}
	}
}


func SubscribeJSON[T any](
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	queueType QueueType,
	handler func(T),
) error {
	channel, queue, _ := DeclareAndBind(conn, exchange, queueName, key, queueType)
	deliveries, _ := channel.Consume(queue.Name, "", false, false, false, false, nil)
	// Start a goroutine to process the deliveries
	//var wg sync.WaitGroup
	// wg.Add(1)
	go StartConsumer(deliveries, handler)

	// Optionally, wait for the goroutine to finish (or handle the wait elsewhere)
	// wg.Wait()

	return nil
}


func PublishJSON[T any](ch *amqp.Channel, exchange, key string, val T) error {
	valueJson, err := json.Marshal(val)
	if err != nil {
		fmt.Println("Error marshalling to JSON:", err)
		return err
	}

	return ch.PublishWithContext(context.Background(), exchange, key, false, false, amqp.Publishing{
		ContentType: "application/json",
		Body: valueJson,
	})
}

func DeclareAndBind(
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	simpleQueueType QueueType, // an enum to represent "durable" or "transient"
) (*amqp.Channel, amqp.Queue, error) {
	channel, err := conn.Channel()
	if err != nil {
		return nil, amqp.Queue{}, err
	}
	queue, err := channel.QueueDeclare(queueName, simpleQueueType == Durable, simpleQueueType == Transient, 
		simpleQueueType == Transient, false, nil)
	fmt.Println(queue.Name)
	if err != nil {
		return nil, amqp.Queue{}, err
	}
	err = channel.QueueBind(queueName, key, exchange, false, nil)
	if err != nil {
		return nil, amqp.Queue{}, err
	}
	return channel, queue, err
	
}