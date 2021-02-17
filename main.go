package main

import (
	"encoding/json"
	"log"
	pool "./pool"
	"os"

	"github.com/streadway/amqp"
)

type Config struct {
	Queue int
}


func main() {
	config := getConfiguration()
	getter(config.Queue)
}

func getConfiguration() *Config {
	file, _ := os.Open("config.json")
	decoder := json.NewDecoder(file)
	config := new(Config)
	err := decoder.Decode(&config)
	failOnError(err, "Fail to decoding")
	return config
}

func getter(countQueue int) {
	conn, err := amqp.Dial("amqp://guest:guest@localhost/")
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()
	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	new_q, err := ch.QueueDeclare(
		"hello", // name
		false,   // durable
		false,   // delete when unused
		false,   // exclusive
		false,   // no-wait
		nil,     // arguments
	)
	failOnError(err, "Failed to declare a queue")
	num, err := ch.Consume(
		new_q.Name, // queue
		"",         // consumer
		true,       // auto-ack
		false,      // exclusive
		false,      // no-local
		false,      // no-wait
		nil,        // args
	)
	failOnError(err, "Failed to register a consumer")

	p := pool.NewPool(countQueue)

	p.Run()
	forever := make(chan bool)
	go func() {
		counter := 1
		for d := range num {
			p.JobQueue <- pool.Job{
				ID:			counter,
				Resources: 	string(d.Body),
			}
		}
	}()
	<-forever
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}
