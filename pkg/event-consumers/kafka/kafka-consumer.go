/*
Copyright (c) 2016-2017 Bitnami

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package kafka

import (
	"context"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Shopify/sarama"
	backoff "github.com/cenkalti/backoff/v4"
	"github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"

	"github.com/kubeless/kafka-trigger/pkg/utils"
)

var (
	stopM      map[string]chan struct{}
	stoppedM   map[string]chan struct{}
	consumerM  map[string]bool
	brokers    string
	maxBackOff time.Duration
	config     *sarama.Config
)

const clientID = "kubeless-kafka-trigger-controller"
const defaultBrokers = "kafka.kubeless:9092"
const kafkatriggersNamespace = "kafkatriggers.kubeless.io"

func init() {
	stopM = make(map[string]chan struct{})
	stoppedM = make(map[string]chan struct{})
	consumerM = make(map[string]bool)

	if os.Getenv("KUBELESS_LOG_LEVEL") == "DEBUG" {
		logrus.SetLevel(logrus.DebugLevel)
	}

	sarama.Logger = logrus.StandardLogger()

	brokers = os.Getenv("KAFKA_BROKERS")
	if brokers == "" {
		brokers = defaultBrokers
	}

	if s := os.Getenv("BACKOFF_INTERVAL"); len(s) > 0 {
		if d, err := time.ParseDuration(s); err == nil {
			maxBackOff = d
		} else {
			logrus.Errorf("Failed to parse maximum back off interval BACKOFF_INTERVAL: %v", err)
		}
	}

	config = sarama.NewConfig()
	config.ClientID = clientID
	config.Version = sarama.V0_10_2_0 // Min supported version for consumer groups.
	config.Consumer.Return.Errors = true

	var err error

	if enableTLS, _ := strconv.ParseBool(os.Getenv("KAFKA_ENABLE_TLS")); enableTLS {
		config.Net.TLS.Enable = true
		config.Net.TLS.Config, err = GetTLSConfiguration(os.Getenv("KAFKA_CACERTS"), os.Getenv("KAFKA_CERT"), os.Getenv("KAFKA_KEY"), os.Getenv("KAFKA_INSECURE"))
		if err != nil {
			logrus.Fatalf("Failed to set tls configuration: %v", err)
		}
	}
	if enableSASL, _ := strconv.ParseBool(os.Getenv("KAFKA_ENABLE_SASL")); enableSASL {
		config.Net.SASL.Enable = true
		config.Net.SASL.User, config.Net.SASL.Password, err = GetSASLConfiguration(os.Getenv("KAFKA_USERNAME"), os.Getenv("KAFKA_PASSWORD"))
		if err != nil {
			logrus.Fatalf("Failed to set SASL configuration: %v", err)
		}
	}
}

// createConsumerProcess gets messages to a Kafka topic from the broker and send the payload to function service
func createConsumerProcess(topic, funcName, ns, consumerGroupID string, clientset kubernetes.Interface, stopchan, stoppedchan chan struct{}) {
	defer close(stoppedchan)

	group, err := sarama.NewConsumerGroup(strings.Split(brokers, ","), consumerGroupID, config)
	if err != nil {
		logrus.Fatalf("Create consumer group (brokers = %v topic = %v namespace = %v function = %v consumerID = %v): %v", brokers, topic, ns, funcName, consumerGroupID, err)
	}
	defer func() {
		if err := group.Close(); err != nil {
			logrus.Errorf("Close consumer group (brokers = %v topic = %v namespace = %v function = %v consumerID = %v): %v", brokers, topic, ns, funcName, consumerGroupID, err)
		}
	}()

	funcPort, err := utils.GetFunctionPort(clientset, ns, funcName)
	if err != nil {
		logrus.Fatalf("Cannot get function port (namespace = %v function = %v): %v", ns, funcName, err)
	}

	ready := make(chan struct{})
	consumer := NewConsumer(funcName, funcPort, ns, clientset, ready, maxBackOff)
	errchan := group.Errors()

	go func() {
		for err := range errchan {
			logrus.Errorf("Consumer group (brokers = %v topic = %v namespace = %v function = %v consumerID = %v): %v", brokers, topic, ns, funcName, consumerGroupID, err)
		}
	}()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			if err := group.Consume(ctx, []string{topic}, consumer); err != nil {
				logrus.Errorf("Consumer group consuming (brokers = %v topic = %v namespace = %v function = %v consumerID = %v): %v", brokers, topic, ns, funcName, consumerGroupID, err)
			}
			if ctx.Err() != nil {
				return
			}
			consumer.Reset()
		}
	}()

	select {
	case <-ready:
	case <-stopchan:
		cancel()
	}

	wg.Wait()
}

// CreateKafkaConsumer creates a goroutine that subscribes to Kafka topic
func CreateKafkaConsumer(triggerObjName, funcName, ns, topic string, clientset kubernetes.Interface) error {
	consumerID := generateUniqueConsumerGroupID(triggerObjName, funcName, ns, topic)
	if consumerM[consumerID] {
		logrus.Debugf("Creating consumer (namespace = %v function = %v trigger = %v topic = %v): already exists, skipping", ns, funcName, triggerObjName, topic)
		return nil
	}

	logrus.Debugf("Creating consumer (namespace = %v function = %v trigger = %v topic = %v)", ns, funcName, triggerObjName, topic)
	stopM[consumerID] = make(chan struct{})
	stoppedM[consumerID] = make(chan struct{})
	go createConsumerProcess(topic, funcName, ns, consumerID, clientset, stopM[consumerID], stoppedM[consumerID])
	consumerM[consumerID] = true

	return nil
}

// DeleteKafkaConsumer deletes goroutine created by CreateKafkaConsumer
func DeleteKafkaConsumer(triggerObjName, funcName, ns, topic string) error {
	consumerID := generateUniqueConsumerGroupID(triggerObjName, funcName, ns, topic)
	if !consumerM[consumerID] {
		logrus.Debugf("Stopping consumer (namespace = %v function = %v trigger = %v topic = %v): does not exist, skipping", ns, funcName, triggerObjName, topic)
		return nil
	}

	logrus.Debugf("Stopping consumer (namespace = %v function = %v trigger = %v topic = %v)", ns, funcName, triggerObjName, topic)
	close(stopM[consumerID])
	<-stoppedM[consumerID]
	consumerM[consumerID] = false
	logrus.Debugf("Stopped consumer (namespace = %v function = %v trigger = %v topic = %v)", ns, funcName, triggerObjName, topic)

	return nil
}

func generateUniqueConsumerGroupID(triggerObjName, funcName, ns, topic string) string {
	return ns + "_" + triggerObjName + "_" + funcName + "_" + topic
}

// Consumer represents a Sarama consumer group consumer.
type Consumer struct {
	funcName  string
	funcPort  int
	ns        string
	clientset kubernetes.Interface
	ready     chan struct{}
	backoff   time.Duration
}

// NewConsumer returns new consumer.
func NewConsumer(funcName string, funcPort int, ns string, clientset kubernetes.Interface, ready chan struct{}, backoff time.Duration) *Consumer {
	return &Consumer{
		clientset: clientset,
		funcName:  funcName,
		funcPort:  funcPort,
		ns:        ns,
		ready:     ready,
		backoff:   backoff,
	}
}

// Reset resets the consumer for new session.
func (c *Consumer) Reset() {
	c.ready = make(chan struct{})
}

// Setup is run at the beginning of a new session, before ConsumeClaim.
func (c *Consumer) Setup(sarama.ConsumerGroupSession) error {
	// Mark the consumer as ready.
	close(c.ready)
	return nil
}

// Cleanup is run at the end of a session, once all ConsumeClaim goroutines have exited.
func (c *Consumer) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

// ConsumeClaim must start a consumer loop of ConsumerGroupClaim's Messages().
func (c *Consumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	b := getBackOff(c.backoff)

	for msg := range claim.Messages() {
		req, err := utils.GetHTTPReq(c.funcName, c.funcPort, msg, c.ns, kafkatriggersNamespace)
		if err != nil {
			logrus.Errorf("Unable to elaborate request (namespace = %v function = %v topic = %v partition = %v offset = %v): %v", c.ns, c.funcName, msg.Topic, msg.Partition, msg.Offset, err)
			continue
		}

		err = utils.SendMessage(req)
		session.MarkMessage(msg, "")

		if err != nil {
			d := b.NextBackOff()
			logrus.Errorf("Failed to send message (namespace = %v function = %v topic = %v partition = %v offset = %v): %v: backing off for %v", c.ns, c.funcName, msg.Topic, msg.Partition, msg.Offset, err, d)
			time.Sleep(d)
			continue
		}

		logrus.Infof("Message sent successfully (namespace = %v function = %v topic = %v partition = %v offset = %v)", c.ns, c.funcName, msg.Topic, msg.Partition, msg.Offset)
		b.Reset()
	}
	return nil
}

type backOff interface {
	NextBackOff() time.Duration
	Reset()
}

type noopBackOff struct{}

func (noopBackOff) NextBackOff() time.Duration { return 0 }
func (noopBackOff) Reset()                     {}

func getBackOff(maxBackOff time.Duration) backOff {
	if maxBackOff < 1 {
		return noopBackOff{}
	}

	b := backoff.NewExponentialBackOff()
	b.MaxElapsedTime = 0 // ... so that b.NextBackOff() never returns backoff.Stop.
	b.MaxInterval = maxBackOff
	return b
}
