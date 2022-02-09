package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/streadway/amqp"
)

var connection *amqp.Connection
var channel *amqp.Channel

func homePage(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Welcome to the HomePage!")
	fmt.Println("Endpoint Hit: homePage")
}

func productPrice(w http.ResponseWriter, r *http.Request) {

	body, err := io.ReadAll(r.Body)
	failOnError(err, "Failed to read the body content")

	query := r.URL.Path
	list := strings.Split(query, "/")

	failOnError(err, "Failed to read the body content")

	if body != nil {
		publish("queue-1", []byte(list[2]))
	}
}

func handleRequests() {
	http.HandleFunc("/", homePage)
	http.HandleFunc("/productPrice/", productPrice)

	log.Fatal(http.ListenAndServe(":10000", nil))
}

func connect() *amqp.Connection {
	println("connecting to rabbit")
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	failOnError(err, "Failed to connect to RabbitMQ")
	// defer conn.Close()
	return conn

}

func createChannel(name string) *amqp.Channel {
	ch, err := connection.Channel()
	failOnError(err, "Failed to open a channel")
	// defer ch.Close()
	return ch
}

func declareQueue(name string) amqp.Queue {
	q, err := channel.QueueDeclare(
		name,  // name
		false, // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)
	failOnError(err, "Failed to declare a queue")
	return q

}

func publish(routingKey string, body []byte) {
	err := channel.Publish(
		"",         // exchange
		routingKey, // routing key
		false,      // mandatory
		false,      // immediate

		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(body),
		})
	failOnError(err, "Failed to publish a message")
}

func main() {
	connection = connect()
	channel = createChannel("channel-1")

	declareQueue("queue-1")

	handleRequests()
	defer channel.Close()
	defer connection.Close()
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}
