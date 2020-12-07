package consumer

import (
	"github.com/Shopify/sarama"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"strings"
	"testing"
)

func TestKafka(test *testing.T) {
	sources := []string{"110.202.20.10:39092"}
	client, err := sarama.NewClient(sources, nil)
	if err != nil {
		log.Error(err)
	}

	topics, err := client.Topics()
	if err != nil {
		log.Error(err)
	}

	err = ioutil.WriteFile("topics.txt", []byte(strings.Join(topics, "\n")), os.ModePerm)
	if err != nil {
		log.Error(err)
	}
	log.Infof("%d topics", len(topics))

	c, err := sarama.NewConsumerFromClient(client)
	if err != nil {
		return
	}

	if len(topics) != 0 {
		partitions, err := c.Partitions(topics[0])
		if err != nil {
			return
		}

		// [0 1 2 3 4 5 6 7 8 9 10 11 12 13 14 15 16]
		log.Info(partitions)
	}
}
