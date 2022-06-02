package websocket

import (
	"github.com/streadway/amqp"
)

var Buf = make(chan []byte, 64)

func Send(msg string) {
	Buf <- []byte(msg)
}

func Run() {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()
	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()
	q, err := ch.QueueDeclare(
		"hello", // name
		false,   // durable
		false,   // delete when unused
		false,   // exclusive
		false,   // no-wait
		nil,     // arguments
	)
	failOnError(err, "Failed to declare a queue")

	pool := make(chan []byte, 64)
	for {
		select {
		case output := <-Buf:
			pool <- output
			go work(pool, q, ch)
		}
	}
}

func work(pool <-chan []byte, q amqp.Queue, ch *amqp.Channel) {
	for msg := range pool {
		err := ch.Publish(
			"",     // exchange
			q.Name, // routing key
			false,  // mandatory
			false,  // immediate
			amqp.Publishing{
				ContentType: "text/plain",
				Body:        msg,
			})

		failOnError(err, "Failed to publish a message")
	}
}
