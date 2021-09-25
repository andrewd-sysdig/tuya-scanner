package agent

import (
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

// Agent runs an mqtt client
type Agent struct {
	client        mqtt.Client
	clientID      string
	terminated    bool
	subscriptions []subscription
}

type subscription struct {
	topic   string
	handler mqtt.MessageHandler
}

// Close agent
func (a *Agent) Close() {
	a.client.Disconnect(250)
	a.terminated = true
}

// Subscribe to topic
func (a *Agent) Subscribe(topic string, handler mqtt.MessageHandler) (err error) {
	token := a.client.Subscribe(topic, 2, handler)
	if !token.WaitTimeout(2 * time.Second) {
		return errors.New("Subscribe timout")
	}
	if token.Error() != nil {
		return token.Error()
	}
	// log.WithFields(log.Fields{
	// 	"topic": topic}).Info("Subscribe")
	a.subscriptions = append(a.subscriptions, subscription{topic, handler})

	return nil
}

// Publish things
func (a *Agent) Publish(topic string, retain bool, payload []byte) (err error) {
	token := a.client.Publish(topic, 2, retain, payload)
	if !token.WaitTimeout(2 * time.Second) {
		return errors.New("Publish timout")
	}
	if token.Error() != nil {
		return token.Error()
	}
	return nil
}

// NewAgent creates an agent
func NewAgent(address string, port int, user string, password string, clientID string) (a *Agent) {
	a = &Agent{}
	a.clientID = clientID

	// mqtt.DEBUG = log.New(os.Stdout, "", 0)
	// mqtt.ERROR = log.New(os.Stdout, "", 0)
	opts := mqtt.NewClientOptions()
	opts.AddBroker(fmt.Sprintf("tcp://%s:%d", address, port))
	opts.SetClientID(clientID)
	if user != "" {
		opts.SetUsername(user)
	}
	if password != "" {
		opts.SetPassword(password)
	}
	opts.SetKeepAlive(5 * time.Second)
	opts.SetPingTimeout(5 * time.Second)
	opts.ConnectRetry = true
	opts.AutoReconnect = true
	opts.OnReconnecting = func(mqtt.Client, *mqtt.ClientOptions) {
		log.Println("MQTT Broker - Attempting to reconnect")
	}
	opts.OnConnectionLost = func(c mqtt.Client, err error) {
		log.Printf("MQTT Broker - Lost connection: %v\n", err)
	}
	opts.OnConnect = func(c mqtt.Client) {
		log.Println("MQTT Broker - Connected to", fmt.Sprintf("tcp://%s:%d", address, port))

		//Subscribe here, otherwise after connection lost,
		//you may not receive any message
		for _, s := range a.subscriptions {
			if token := c.Subscribe(s.topic, 2, s.handler); token.Wait() && token.Error() != nil {
				log.Println("Can't Subscribe", token.Error())
				os.Exit(1)
			}
			log.Println("Resubscribe - Topic:", s.topic)
		}
	}
	a.client = mqtt.NewClient(opts)

	return a
}

// Connect opens a new connection
func (a *Agent) Connect() (err error) {
	token := a.client.Connect()
	if !token.WaitTimeout(2 * time.Second) {
		return errors.New("open timeout")
	}
	if token.Error() != nil {
		return token.Error()
	}
	return
}

func (a *Agent) IsTerminated() bool {
	return a.terminated
}
