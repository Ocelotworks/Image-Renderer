package main

import (
	"encoding/json"
	"fmt"
	"github.com/getsentry/sentry-go"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/shirou/gopsutil/cpu"
	"github.com/streadway/amqp"
	"gl.ocelotworks.com/ocelotbotv5/image-renderer/entity"
	"golang.org/x/image/webp"
	"image"
	"log"
	"net/http"
	"os"
)

var messagesProcessed = promauto.NewCounter(prometheus.CounterOpts{
	Namespace: "image_renderer",
	Name:      "messages_processed",
	Help:      "The total number of processed messages",
})

func main() {

	_ = sentry.Init(sentry.ClientOptions{})

	image.RegisterFormat("webp", "RIFF", webp.Decode, webp.DecodeConfig)

	conn, exception := amqp.Dial(os.Getenv("RABBIT_URL"))

	if exception != nil {
		sentry.CaptureException(exception)
		log.Fatalf("Failed to connect to RabbitMQ: %s", exception)
	}

	channel, exception := conn.Channel()

	if exception != nil {
		sentry.CaptureException(exception)
		log.Fatalf("Failed to open channel: %s", exception)
	}

	priority := 0

	cpuInfo, exception := cpu.Info()

	if exception != nil {
		sentry.CaptureException(exception)
		log.Println("Couldn't get CPU info:", exception)
	} else {
		priority = int(cpuInfo[0].Mhz / 100)
	}

	log.Println("Consumer priority: ", priority)

	_, exception = channel.QueueDeclare("imageProcessor", false, false, false, false, map[string]interface{}{
		"x-message-ttl": 60000,
		"x-priority":    priority,
	})

	if exception != nil {
		sentry.CaptureException(exception)
		log.Fatalf("Failed to declare queue: %s", exception)
	}

	_ = channel.Qos(4, 0, true)

	messages, exception := channel.Consume("imageProcessor", "", false, false, false, false, nil)

	if exception != nil {
		sentry.CaptureException(exception)
		log.Fatalf("Failed to consume queue: %s", exception)
	}

	fmt.Println("Ready!")

	forever := make(chan bool)

	go func() {
		for messageData := range messages {
			messagesProcessed.Inc()
			go processMessage(messageData, channel)
		}
	}()

	http.Handle("/metrics", promhttp.Handler())
	_ = http.ListenAndServe(":2112", nil)

	<-forever
}

func processMessage(messageData amqp.Delivery, channel *amqp.Channel) {
	log.Printf("Received Message: %s", messageData.Body)
	imageRequest := entity.ImageRequest{}
	exception := json.Unmarshal(messageData.Body, &imageRequest)
	if exception != nil {
		log.Printf("Malformed message: %s", exception)
	} else {
		exception := reply(channel, messageData, ProcessImage(&imageRequest))
		if exception != nil {
			sentry.CaptureException(exception)
			log.Println("Unable to send response: ", exception)
			exception = reply(channel, messageData, &entity.ImageResult{Error: "reply"})
			if exception != nil {
				sentry.CaptureException(exception)
				log.Fatalln("Unable to send error message response: ", exception)
			}
		}
	}
	_ = messageData.Ack(false)
}

func reply(channel *amqp.Channel, recipient amqp.Delivery, result *entity.ImageResult) error {
	output, exception := json.Marshal(result)
	if exception != nil {
		return exception
	}
	return channel.Publish("", recipient.ReplyTo, false, false, amqp.Publishing{
		CorrelationId: recipient.CorrelationId,
		Body:          output,
	})
}
