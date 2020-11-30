package main

import (
	"encoding/json"
	"github.com/streadway/amqp"
	"gl.ocelotworks.com/ocelotbotv5/image-renderer/entity"
	"log"
	"os"
)

func main() {

	conn, exception := amqp.Dial(os.Getenv("RABBIT_URL"))

	if exception != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %s", exception)
	}

	channel, exception := conn.Channel()

	if exception != nil {
		log.Fatalf("Failed to open channel: %s", exception)
	}

	_, exception = channel.QueueDeclare("imageProcessor", false, false, false, false, map[string]interface{}{
		"x-message-ttl": 60000,
	})

	if exception != nil {
		log.Fatalf("Failed to declare queue: %s", exception)
	}

	messages, exception := channel.Consume("imageProcessor", "", false, false, false, false, nil)

	if exception != nil {
		log.Fatalf("Failed to consume queue: %s", exception)
	}

	forever := make(chan bool)

	go func() {
		for messageData := range messages {
			log.Printf("Received Message: %s", messageData.Body)
			imageRequest := entity.ImageRequest{}
			exception = json.Unmarshal(messageData.Body, &imageRequest)
			if exception != nil {
				log.Printf("Malformed message: %s", exception)
			} else {
				ProcessImage(&imageRequest)
			}
			_ = messageData.Ack(false)
		}
	}()

	<-forever
}
