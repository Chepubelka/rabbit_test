package main

import (
	"log"
	"net/http"
	"strconv"
	"encoding/json"
	"path"
	"html/template"
	"github.com/gorilla/mux"
	"github.com/streadway/amqp"
)

func main() {
	handleRequests()
}

func handleRequests() {
	myRouter := mux.NewRouter().StrictSlash(true)
	myRouter.HandleFunc("/counter/{count}", counter)
	myRouter.HandleFunc("/render/{count}", render)
	log.Fatal(http.ListenAndServe(":8080", myRouter))
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
	w.Header().Add("Content-Type", "application/json")
	vars := mux.Vars(r)
	count := vars["count"]
	countInt, _ := strconv.Atoi(count)
	result := sender(countInt)
	response, _ := json.Marshal(result)
	w.Write(response)
}

func sender(countInt int) []string{
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
			"",     // consumer
			true,   // auto-ack
			false,  // exclusive
			false,  // no-local
			false,  // no-wait
			nil,    // args
		  )
		  failOnError(err, "Failed to register a consumer")
		  var responseArray []string
		  func() {
			for i := range num {
				responseArray = append(responseArray, string(i.Body))
				if (len(responseArray) == countInt) {
					break
				}
			}
		  }()
		return responseArray
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}
