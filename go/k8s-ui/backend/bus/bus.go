package bus

import (
	"github.com/streadway/amqp"
)

const (
	RoutingKeyRequest = "request"
)

var DefaultBus *Bus

type Bus struct {
	Name    string
	Url     string
	Conn    *amqp.Connection
	Channel *amqp.Channel
}

func NewBus(url string) (*Bus, error) {
	bus := &Bus{Name: "k8s-ui", Url: url}

	conn, err := amqp.Dial(bus.Url)
	if err != nil {
		return nil, err
	}
	bus.Conn = conn

	channel, err := conn.Channel()
	if err != nil {
		return nil, err
	}
	bus.Channel = channel

	if err := bus.Init(); err != nil {
		return nil, err
	}

	return bus, nil
}

func (bus *Bus) Init() error {
	// Exchange
	if err := bus.Channel.ExchangeDeclare(bus.Name, amqp.ExchangeDirect,
		true,
		false,
		false,
		false,
		nil); err != nil {
		return err
	}

	return nil
}
