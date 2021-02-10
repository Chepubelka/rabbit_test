package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/streadway/amqp"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func main() {
	handleRequests()
}

func handleRequests() {
	myRouter := mux.NewRouter().StrictSlash(true)
	myRouter.HandleFunc("/counter", counter)
	myRouter.HandleFunc("/render", render)
	log.Fatal(http.ListenAndServe(":8081", myRouter))
}

func render(w http.ResponseWriter, r *http.Request) {
	fp := path.Join("views", "index.html")
	tmpl, err := template.ParseFiles(fp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := tmpl.Execute(w, ""); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func counter(w http.ResponseWriter, r *http.Request) {
	connect, _ := upgrader.Upgrade(w, r, nil) // error ignored for sake of simplicity
	for {
		// Read message from browser
		msgType, msg, err := connect.ReadMessage()
		if err != nil {
			return
		}
		countInt, _ := strconv.Atoi(string(msg))
		sender(countInt)
		conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
		failOnError(err, "Failed to connect to RabbitMQ")
		defer conn.Close()
		ch, err := conn.Channel()
		failOnError(err, "Failed to open a channel")
		defer ch.Close()
		new_q, err := ch.QueueDeclare(
			"test1", // name
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
		for i := range num {
			fmt.Printf("%s sent: %s\n", connect.RemoteAddr(), i.Body)
			// Write message back to browser
			if err = connect.WriteMessage(msgType, i.Body); err != nil {
				return
			}
		}
	}
}

func sender(countInt int){
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()
	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()
	q, err := ch.QueueDeclare(
		"test", // name
		false,  // durable
		false,  // delete when unused
		false,  // exclusive
		false,  // no-wait
		nil,    // arguments
	)

	for i := 0; i < countInt; i++ {
		err = ch.Publish(
			"",     // exchange
			q.Name, // routing key
			false,  // mandatory
			false,  // immediate
			amqp.Publishing{
				ContentType: "text/plain",
				Body:        []byte(strconv.Itoa(i)),
			})
		failOnError(err, "Failed to publish a message")
	}
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}
