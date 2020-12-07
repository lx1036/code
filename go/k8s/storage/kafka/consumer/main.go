package main

import (
	"flag"
	"fmt"
	"github.com/Shopify/sarama"
	"k8s-lx1036/k8s/storage/kafka/tls"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
)

// https://github.com/Shopify/sarama/blob/master/tools/kafka-console-consumer/kafka-console-consumer.go

var (
	brokerList    = flag.String("brokers", os.Getenv("KAFKA_PEERS"), "The comma separated list of brokers in the Kafka cluster")
	topic         = flag.String("topic", "", "REQUIRED: the topic to consume")
	partitions    = flag.String("partitions", "all", "The partitions to consume, can be 'all' or comma-separated numbers")
	offset        = flag.String("offset", "newest", "The offset to start with. Can be `oldest`, `newest`")
	verbose       = flag.Bool("verbose", false, "Whether to turn on sarama logging")
	tlsEnabled    = flag.Bool("tls-enabled", false, "Whether to enable TLS")
	tlsSkipVerify = flag.Bool("tls-skip-verify", false, "Whether skip TLS server cert verification")
	tlsClientCert = flag.String("tls-client-cert", "", "Client cert for client authentication (use with -tls-enabled and -tls-client-key)")
	tlsClientKey  = flag.String("tls-client-key", "", "Client key for client authentication (use with tls-enabled and -tls-client-cert)")

	bufferSize = flag.Int("buffer-size", 256, "The buffer size of the message channel.")

	logger = log.New(os.Stderr, "", log.LstdFlags)
)

// https://github.com/Shopify/sarama/blob/master/tools/kafka-console-consumer/kafka-console-consumer.go
// go run . --brokers= --topic=
func main() {
	flag.Parse()

	if *brokerList == "" {
		printUsageErrorAndExit("You have to provide -brokers as a comma-separated list, or set the KAFKA_PEERS environment variable.")
	}

	if *topic == "" {
		printUsageErrorAndExit("-topic is required")
	}

	if *verbose {
		sarama.Logger = logger
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

	config := sarama.NewConfig()
	if *tlsEnabled {
		tlsConfig, err := tls.NewConfig(*tlsClientCert, *tlsClientKey)
		if err != nil {
			printErrorAndExit(69, "Failed to create TLS config: %s", err)
		}

		config.Net.TLS.Enable = true
		config.Net.TLS.Config = tlsConfig
		config.Net.TLS.Config.InsecureSkipVerify = *tlsSkipVerify
	}

	c, err := sarama.NewConsumer(strings.Split(*brokerList, ","), config)
	if err != nil {
		printErrorAndExit(69, "Failed to start consumer: %s", err)
	}

	partitionList, err := getPartitions(c)
	if err != nil {
		printErrorAndExit(69, "Failed to get the list of partitions: %s", err)
	}

	var (
		messages = make(chan *sarama.ConsumerMessage, *bufferSize)
		closing  = make(chan struct{})
		wg       sync.WaitGroup
	)

	go func() {
		signals := make(chan os.Signal, 1)
		signal.Notify(signals, syscall.SIGTERM, os.Interrupt)
		<-signals
		logger.Println("Initiating shutdown of consumer...")
		close(closing)
	}()

	// 每一个partition，一个partitionConsumer来消费。
	// Partition能水平扩展客户端的读写性能，是高吞吐量的保障，就是磁盘上存放数据的文件夹。
	// Partition下还可以细分成Segment，每个Segment分别对应两个文件：一个是以.index结尾的索引文件，另一个是以.log结尾的数据文件，且两个文件的文件名完全相同，
	// 所有的Segment均存在于所属Partition的目录下。
	for _, partition := range partitionList {
		partitionConsumer, err := c.ConsumePartition(*topic, partition, initialOffset)
		if err != nil {
			printErrorAndExit(69, "Failed to start consumer for partition %d: %s", partition, err)
		}

		go func(pc sarama.PartitionConsumer) {
			<-closing
			pc.AsyncClose()
		}(partitionConsumer)

		wg.Add(1)
		go func(pc sarama.PartitionConsumer) {
			defer wg.Done()
			for message := range pc.Messages() {
				messages <- message
			}
		}(partitionConsumer)
	}

	go func() {
		for msg := range messages {
			fmt.Printf("Partition:\t%d\n", msg.Partition)
			fmt.Printf("Offset:\t%d\n", msg.Offset)
			fmt.Printf("Key:\t%s\n", string(msg.Key))
			fmt.Printf("Value:\t%s\n", string(msg.Value))
			fmt.Println()
		}
	}()

	wg.Wait()
	logger.Println("Done consuming topic", *topic)
	close(messages)

	if err := c.Close(); err != nil {
		logger.Println("Failed to close consumer: ", err)
	}
}

func printErrorAndExit(code int, format string, values ...interface{}) {
	fmt.Fprintf(os.Stderr, "ERROR: %s\n", fmt.Sprintf(format, values...))
	fmt.Fprintln(os.Stderr)
	os.Exit(code)
}

func printUsageErrorAndExit(format string, values ...interface{}) {
	fmt.Fprintf(os.Stderr, "ERROR: %s\n", fmt.Sprintf(format, values...))
	fmt.Fprintln(os.Stderr)
	fmt.Fprintln(os.Stderr, "Available command line options:")
	flag.PrintDefaults()
	os.Exit(64)
}

func getPartitions(c sarama.Consumer) ([]int32, error) {
	if *partitions == "all" {
		return c.Partitions(*topic)
	}

	tmp := strings.Split(*partitions, ",")
	var pList []int32
	for i := range tmp {
		val, err := strconv.ParseInt(tmp[i], 10, 32)
		if err != nil {
			return nil, err
		}
		pList = append(pList, int32(val))
	}

	return pList, nil
}
