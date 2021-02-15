package main

import (
	"fmt"
	"log"
	"net/http"
	"net/http/pprof"
	"net/url"
	"os"
	"strconv"

	"os/signal"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
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
	log.Fatal(http.ListenAndServe(":8081", myRouter))
}

func main() {
	handleRequests()
}

func startQueue(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	count := vars["count"]
	countInt, _ := strconv.Atoi(count)
	//name_log := generateLogs(countInt)
	sender(countInt)
	getter()
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
			if IntI%2 == 0 {
				fmt.Print(string(i.Body) + "\n")
				file.WriteString(string(i.Body) + "\n")
			}
		}
	}()
	<-forever
}

func sender(count int) {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	u := url.URL{Scheme: "ws", Host: "192.168.2.50:3000", Path: "/queue/receive"}
	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer c.Close()

	go func() {
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				log.Println("read:", err)
				return
			}
			log.Printf("recv: %s", message)
		}
	}()
	done := make(chan struct{})
	countChan := make(chan string)
	countString := strconv.Itoa(count)
	for {
		select {
		case <-done:
			return
		case <-countChan:
			err := c.WriteMessage(websocket.TextMessage, []byte("send " + countString))
			if err != nil {
				log.Println("write:", err)
				return
			}
			return
		case <-interrupt:
			log.Println("interrupt")

			// Cleanly close the connection by sending a close message and then
			// waiting (with timeout) for the server to close the connection.
			err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				log.Println("write close:", err)
				return
			}
		}
	}
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}
