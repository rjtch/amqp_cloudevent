package main

import (
	"context"
	"github.com/Azure/go-amqp"
	ceamqp "github.com/cloudevents/sdk-go/protocol/amqp/v2"
	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/google/uuid"
	"log"
	"net/url"
	"os"
	"strings"
	"time"
)

const (
	count = 10
)

// Parse AM!P_URL from env variable and return URl
func loadConfig() (server, node string, opts []ceamqp.Option) {
	env := os.Getenv("AMQP_URL")
	if env == "" {
		env = "amqp://guest:guest@localhost:5672/test"
	}

	u, err := url.Parse(env)
	if err != nil {
		log.Fatal(err)
	}

	if u.User != nil {
		user := u.User.Username()
		password, _ := u.User.Password()
		opts = append(opts, ceamqp.WithConnOpt(amqp.ConnSASLPlain(user, password)))
	}

	return env, strings.TrimPrefix(u.Path, "/"), opts
}

//basic data structure to be sent
type data struct {
	Sequence int    `json:"id"`
	Message  string `json:"message"`
}

func main() {
	host, node, opts := loadConfig()
	p, err := ceamqp.NewProtocol(host, node, []amqp.ConnOption{}, []amqp.SessionOption{}, opts...)
	if err != nil {
		log.Fatalf("Failed to create amqp protocol: %v", err)
	}

	// Close the connection when finished
	defer p.Close(context.Background())

	// Create new client from the given protocol
	c, err := cloudevents.NewClient(p)
	if err != nil {
		log.Fatal("Failed to create amqp-client: ", err)
	}

	// Send count=10 events from the client
	for i := 0; i < count*1000; i++ {
		event := cloudevents.NewEvent()
		event.SetID(uuid.New().String())
		event.SetSource("/")
		event.SetTime(time.Now())
		event.SetType("com.events.rjtch")

		// create data to be sent
		err := event.SetData(cloudevents.ApplicationJSON,
			&data{Sequence: i, Message: "Hello receiver"})
		if err != nil {
			log.Fatal("Failed to set data: ", err)
		}

		// send data created
		result := c.Send(context.Background(), event)
		log.Println("result: ", result)
		if cloudevents.IsUndelivered(result) {
			log.Fatalf("Failed to send: %v", result)
		} else if cloudevents.IsNACK(result) {
			log.Printf("Event not accepted: %v", result)
		}

		log.Println("event: ", i, event)

		time.Sleep(100 * time.Millisecond)
	}
}
