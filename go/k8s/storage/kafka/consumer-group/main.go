package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"

	"github.com/Shopify/sarama"
)

var (
	brokers  = flag.String("brokers", os.Getenv("KAFKA_PEERS"), "The comma separated list of brokers in the Kafka cluster")
	topics   = flag.String("topics", "", "REQUIRED: the topic to consume")
	group    = flag.String("group", "", "Kafka consumer group definition")
	assignor = flag.String("assignor", "range", "Consumer group partition assignment strategy (range, roundrobin, sticky)")
	verbose  = flag.Bool("verbose", false, "Whether to turn on sarama logging")
	offset   = flag.String("offset", "newest", "The offset to start with. Can be `oldest`, `newest`")
	version  = flag.String("version", "2.1.1", "Kafka cluster version")

	tlsEnabled    = flag.Bool("tls-enabled", false, "Whether to enable TLS")
	tlsSkipVerify = flag.Bool("tls-skip-verify", false, "Whether skip TLS server cert verification")
	tlsClientCert = flag.String("tls-client-cert", "", "Client cert for client authentication (use with -tls-enabled and -tls-client-key)")
	tlsClientKey  = flag.String("tls-client-key", "", "Client key for client authentication (use with tls-enabled and -tls-client-cert)")

	bufferSize = flag.Int("buffer-size", 256, "The buffer size of the message channel.")
)

// https://github.com/Shopify/sarama/blob/master/examples/consumergroup/main.go
// go run . --brokers= --topics=test-producer --offset=oldest --group=hello
func main() {
	flag.Parse()

	if len(*brokers) == 0 {
		panic("no Kafka bootstrap brokers defined, please set the -brokers flag")
	}

	if len(*topics) == 0 {
		panic("no topics given to be consumed, please set the -topics flag")
	}

	if len(*group) == 0 {
		panic("no Kafka consumer group defined, please set the -group flag")
	}

	if *verbose {
		sarama.Logger = log.New(os.Stdout, "[sarama] ", log.LstdFlags)
	}

	version, err := sarama.ParseKafkaVersion(*version)
	if err != nil {
		log.Panicf("Error parsing Kafka version: %v", err)
	}
	/**
	 * Construct a new Sarama configuration.
	 * The Kafka cluster version has to be defined before the consumer/producer is initialized.
	 */
	config := sarama.NewConfig()
	config.Version = version

	switch *assignor {
	case "roundrobin":
		config.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategyRoundRobin
	case "range":
		config.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategyRange
	default:
		log.Panicf("Unrecognized consumer group partition assignor: %s", *assignor)
	}

	var initialOffset int64
	switch *offset {
	case "oldest":
		initialOffset = sarama.OffsetOldest
	case "newest":
		initialOffset = sarama.OffsetNewest
	default:
		printUsageErrorAndExit("-offset should be `oldest` or `newest`")
	}
	config.Consumer.Offsets.Initial = initialOffset

	/**
	 * Setup a new Sarama consumer group
	 */
	consumer := Consumer{
		ready: make(chan bool),
	}

	client, err := sarama.NewConsumerGroup(strings.Split(*brokers, ","), *group, config)
	if err != nil {
		log.Panicf("Error creating consumer group client: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			// `Consume` should be called inside an infinite loop, when a
			// server-side rebalance happens, the consumer session will need to be
			// recreated to get the new claims
			if err := client.Consume(ctx, strings.Split(*topics, ","), &consumer); err != nil {
				log.Panicf("Error from consumer: %v", err)
			}
			// check if context was cancelled, signaling that the consumer should stop
			if ctx.Err() != nil {
				return
			}
			consumer.ready = make(chan bool)
		}
	}()

	<-consumer.ready // Await till the consumer has been set up
	log.Println("Sarama consumer up and running!...")

	sigterm := make(chan os.Signal, 1)
	signal.Notify(sigterm, syscall.SIGINT, syscall.SIGTERM)
	select {
	case <-ctx.Done():
		log.Println("terminating: context cancelled")
	case <-sigterm:
		log.Println("terminating: via signal")
	}
	cancel()
	wg.Wait()
	if err = client.Close(); err != nil {
		log.Panicf("Error closing client: %v", err)
	}
}

// Consumer represents a Sarama consumer group consumer
type Consumer struct {
	ready chan bool
}

// Setup is run at the beginning of a new session, before ConsumeClaim
func (consumer *Consumer) Setup(sarama.ConsumerGroupSession) error {
	// Mark the consumer as ready
	close(consumer.ready)
	return nil
}

// Cleanup is run at the end of a session, once all ConsumeClaim goroutines have exited
func (consumer *Consumer) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

// ConsumeClaim must start a consumer loop of ConsumerGroupClaim's Messages().
func (consumer *Consumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	// NOTE:
	// Do not move the code below to a goroutine.
	// The `ConsumeClaim` itself is called within a goroutine, see:
	// https://github.com/Shopify/sarama/blob/master/consumer_group.go#L27-L29
	for message := range claim.Messages() {
		log.Printf("Message claimed: value = %s, timestamp = %v, topic = %s", string(message.Value), message.Timestamp, message.Topic)
		session.MarkMessage(message, "")
	}

	return nil
}

func printUsageErrorAndExit(format string, values ...interface{}) {
	fmt.Fprintf(os.Stderr, "ERROR: %s\n", fmt.Sprintf(format, values...))
	fmt.Fprintln(os.Stderr)
	fmt.Fprintln(os.Stderr, "Available command line options:")
	flag.PrintDefaults()
	os.Exit(64)
}
