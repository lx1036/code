package consumer

import (
	"fmt"
	"github.com/Shopify/sarama"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"strings"
)

type Consumer struct {
	client sarama.Client
	topic  string
}

func NewConsumer() (*Consumer, error) {
	sources := viper.GetStringSlice(fmt.Sprintf("Log.Sources.%s", viper.GetString("source")))
	log.Debugf("use %s with kafka broker list: %s", viper.GetString("source"), strings.Join(sources, ","))
	client, err := sarama.NewClient(sources, nil)
	if err != nil {
		return nil, err
	}

	consumer := &Consumer{
		client: client,
	}
	consumer.topic = consumer.GetTopic()

	return consumer, nil
}

func (consumer *Consumer) GetTopic() string {
	return fmt.Sprintf("k8s_docker_%s", viper.GetString("deployment"))
}

func (consumer *Consumer) GetPartitionOffset() {
	c, err := sarama.NewConsumerFromClient(consumer.client)
	if err != nil {
		return
	}

	partitionIds, err := c.Partitions(consumer.topic)
	if err != nil {
		return
	}

	log.Debug(partitionIds)

}

func (consumer *Consumer) Run(stopCh <-chan struct{}) error {

	return nil
}
