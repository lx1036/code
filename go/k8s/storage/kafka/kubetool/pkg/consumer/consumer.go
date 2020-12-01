package consumer

import (
	"github.com/Shopify/sarama"
)

type Consumer struct {
}

func NewConsumer() (*Consumer, error) {

}

func (consumer *Consumer) Run(stopCh <-chan struct{}) error {

}
