// Package consumer provides an initialiser for a message consumer.
package consumer

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/Shopify/sarama"
	"github.com/companieshouse/chs.go/kafka/consumer"
	"github.com/companieshouse/chs.go/log"
)

// Interface - interface for Kafka consumer
type Interface interface {
	ConsumePartition(partition int32, offset int64) error
	Messages() <-chan *sarama.ConsumerMessage
}

// FactoryInterface -interface for creating the kafka consumer
type FactoryInterface interface {
	Initialise(filingTopic string, brokerAddr []string) Interface
}

// Factory - struct for factory
type Factory struct {
	c *consumer.PartitionConsumer
}

// NewConsumerFactory - create new consumer factory
func NewConsumerFactory() FactoryInterface {
	return &Factory{}
}

// Initialise initialises and returns a new PartitionConsumer.
func (cm *Factory) Initialise(filingTopic string, brokerAddr []string) Interface {
	log.Info("Initialising Partition Consumer", log.Data{"filing topic name": filingTopic})
	c := consumer.NewPartitionConsumer(&consumer.Config{
		Topics:     []string{filingTopic},
		BrokerAddr: brokerAddr,
	})

	go cm.close()
	cm.c = c
	return c
}

func (cm *Factory) close() {
	log.Info("Waiting to close service by sending interrupt signal")
	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	<-ch

	topic := cm.c.Topics[0]
	if err := cm.c.Close(); err != nil {
		log.Error(fmt.Errorf("error closing Partition Consumer: %s", err))
		os.Exit(2)
	}
	if err := cm.c.Client.Close(); err != nil {
		log.Error(fmt.Errorf("error closing Consumer: %s", err))
		os.Exit(2)
	}
	log.Info("Consumer successfully closed", log.Data{"topic": topic})

	os.Exit(0)
}
