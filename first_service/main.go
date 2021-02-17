package main

import (
	"encoding/json"
	//"fmt"
	"log"
	//"math/rand"
	//"net/http"
	//"net/http/pprof"
	"os"

	//"strconv"
	//"time"

	//"github.com/gorilla/mux"
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
	err = ch.Qos(countQueue, 0, false)
	failOnError(err, "Failed to register a consumer")
	file, err := os.Create("new_hello.log")
	failOnError(err, "Error open file")
	defer file.Close()

	




	forever := make(chan bool)
/* 	for i := 0; i < countQueue; i++ {
		go func() {
			fmt.Print("start\n")
			for d := range num {
				r := rand.Intn(5000) + 2000
				strTime := strconv.Itoa(r)
				time.Sleep(time.Duration(r) * time.Millisecond)
				fmt.Print(strTime + "_" + string(d.Body) + "\n")
				file.WriteString(string(d.Body) + "\n")
			}
			os.Exit(1)
		}()
	}
	<-forever */
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}
