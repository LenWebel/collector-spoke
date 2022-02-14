package main

import (
	"context"
	"log"
	"time"

	// "fmt"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/streadway/amqp"
)

func failOnError(err error, msg string) {
	if err != nil {
		// log.Fatalf("%s: %s", msg, err)
		log.Printf("%s: %s", msg, err)
	}
}

func listen() {
	conn, err := amqp.Dial("amqp://guest:guest@rabbitmq-container:5672/")
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	q, err := ch.QueueDeclare(
		"queue-1", // name
		false,     // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)
	failOnError(err, "Failed to declare a queue")

	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	failOnError(err, "Failed to register a consumer")

	forever := make(chan bool)

	go func() {
		for d := range msgs {
			log.Printf("Received a message: %s", d.Body)
		}
	}()

	log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
	<-forever
}

func writeToInflux() {
	bucket := "example-bucket"
	org := "example-org"
	token := "example-token"
	// Store the URL of your InfluxDB instance
	url := "http://localhost:8086"
	// Create new client with default option for server url authenticate by token
	client := influxdb2.NewClient(url, token)
	// User blocking write client for writes to desired bucket
	writeAPI := client.WriteAPIBlocking(org, bucket)
	// Create point using full params constructor
	p := influxdb2.NewPoint("stat",
		map[string]string{"unit": "temperature"},
		map[string]interface{}{"avg": 24.5, "max": 45},
		time.Now())
	// Write point immediately
	err := writeAPI.WritePoint(context.Background(), p)
	// Ensures background processes finishes

	// Get query client
	queryAPI := client.QueryAPI("my-org")
	// Get parser flux query result
	result, err := queryAPI.Query(context.Background(), `from(bucket:"example-bucket")|> range(start: -1h) |> filter(fn: (r) => r._measurement == "stat")`)

	if err == nil {
		log.Println(result)
	}

	client.Close()
}

func main() {
	listen()
}
