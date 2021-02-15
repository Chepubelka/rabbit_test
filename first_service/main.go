package main

import (
	"bufio"
	"fmt"
	"log"
	"net/http"
	"net/http/pprof"
	"os"
	"strconv"
	//s "strings"

	"github.com/gorilla/mux"
	"github.com/streadway/amqp"
)

func handleRequests() {
	myRouter := mux.NewRouter().StrictSlash(true)
	myRouter.HandleFunc("/startQueue/{count}", startQueue)
	myRouter.HandleFunc("/debug/pprof/", pprof.Index)
	myRouter.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	myRouter.HandleFunc("/debug/pprof/profile", pprof.Profile)
	myRouter.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	myRouter.HandleFunc("/debug/pprof/trace", pprof.Trace)
	log.Fatal(http.ListenAndServe(":8080", myRouter))
}

func main() {
	handleRequests()
}

func startQueue(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	count := vars["count"]
	countInt, _ := strconv.Atoi(count)
	name_log := generateLogs(countInt)
	sender(name_log)
	getter()
}

func generateLogs(count int) string {
	var data []string
/* 	data = append(data, "debug\n")
	line := "Barak"
	line = s.Repeat(line, count) + "\n" */
	for i := 1; i <= count; i++ {
		StringI := strconv.Itoa(i) + "\n"
		data = append(data, StringI)
	}
	name_log := "hello.log"
	file, err := os.Create(name_log)
	if err != nil {
		fmt.Println("Unable to create file:", err)
		os.Exit(1)
	}
	defer file.Close()
	for _, val := range data {
		file.WriteString(val)
	}
/* 	w := bufio.NewWriter(file)
	for _, val := range data {
		w.Write([]byte(val))
		failOnError(err, "Fail to write")
	}

	fmt.Println("Ну типа") */
	return name_log
}

func getter() {
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
	forever := make(chan bool)
	file, err := os.Create("new_hello.log")
	failOnError(err, "Error open file")
	defer file.Close()
	//w := bufio.NewWriter(file)
	func() {
		for i := range num {
			IntI, err := strconv.Atoi(string(i.Body))
			failOnError(err, "Fail to write")
			if (IntI % 2 == 0) {
				fmt.Print(string(i.Body) + "\n")
				file.WriteString(string(i.Body) + "\n")
			}
		}
	}()
	<-forever
}

func sender(name_log string) {
	conn, err := amqp.Dial("amqp://guest:guest@192.168.2.50:5672/")
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
	file, err := os.Open(name_log)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer file.Close()
	//var lines []string

	scanner := bufio.NewScanner(file)
/* 	for i := 0; i <= 100; i++ {
		StrI := strconv.Itoa(i)
		err = ch.Publish(
			"",     // exchange
			q.Name, // routing key
			false,  // mandatory
			false,  // immediate
			amqp.Publishing{
				ContentType: "text/plain",
				Body:        []byte(StrI),
			})
		failOnError(err, "Failed to publish a message")
	} */
	for scanner.Scan() {
		//lines = append(lines, scanner.Text())
		err = ch.Publish(
			"",     // exchange
			q.Name, // routing key
			false,  // mandatory
			false,  // immediate
			amqp.Publishing{
				ContentType: "text/plain",
				Body:        []byte(scanner.Text()),
			})
		failOnError(err, "Failed to publish a message")
	}
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}
