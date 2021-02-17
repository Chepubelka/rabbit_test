package main

import (
	"encoding/json"
	"log"
	"net/http"
	"net/http/pprof"
	"os"
	"fmt"
	"math/rand"

	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/streadway/amqp"
)

type Config struct {
	Queue int
}

func handleRequests() {
	myRouter := mux.NewRouter().StrictSlash(true)
	myRouter.HandleFunc("/debug/pprof/", pprof.Index)
	myRouter.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	myRouter.HandleFunc("/debug/pprof/profile", pprof.Profile)
	myRouter.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	myRouter.HandleFunc("/debug/pprof/trace", pprof.Trace)
	log.Fatal(http.ListenAndServe(":8081", myRouter))
}

func main() {
	config := getConfiguration()
	getter(config.Queue)
	//handleRequests()
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
	forever := make(chan bool)
	file, err := os.Create("new_hello.log")
	failOnError(err, "Error open file")
	defer file.Close()
	for i := 0; i < countQueue; i++ {
		go func() {
			for d := range num {
				r := rand.Intn(5000) + 2000
				strTime := strconv.Itoa(r)
				fmt.Print(strTime + "_" + string(d.Body) + "\n")
				time.Sleep(time.Duration(r) * time.Millisecond)
				file.WriteString(string(d.Body) + "\n")
			}
			os.Exit(1)
		}()
	}
	<-forever
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}
