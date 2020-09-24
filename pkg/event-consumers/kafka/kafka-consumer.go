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

	"github.com/Shopify/sarama"
	"github.com/kubeless/kafka-trigger/pkg/utils"
	"github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"
)

var (
	stopM     map[string]chan struct{}
	stoppedM  map[string]chan struct{}
	consumerM map[string]bool
	brokers   string
	config    *sarama.Config
)

func init() {
	stopM = make(map[string]chan struct{})
	stoppedM = make(map[string]chan struct{})
	consumerM = make(map[string]bool)

	if os.Getenv("KUBELESS_LOG_LEVEL") == "DEBUG" {
		logrus.SetLevel(logrus.DebugLevel)
	}

	sarama.Logger = logrus.StandardLogger()

	// Init config
	// taking brokers from env var
	brokers = os.Getenv("KAFKA_BROKERS")
	if brokers == "" {
		brokers = "kafka.kubeless:9092"
	}

	config = sarama.NewConfig()
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
		config.Version = sarama.V0_10_0_0
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
		logrus.Fatalf("Failed to start Kafka consumer brokers = %v topic = %v function = %v consumerID = %v: %v", brokers, topic, funcName, consumerGroupID, err)
	}
	defer func() {
		if err := group.Close(); err != nil {
			logrus.Errorf("Closing Kafka consumer brokers = %v topic = %v function = %v consumerID = %v: %v", brokers, topic, funcName, consumerGroupID, err)
		}
	}()

	logrus.Infof("Started Kafka consumer brokers = %v topic = %v function = %v consumerID = %v", brokers, topic, funcName, consumerGroupID)
	ready := make(chan struct{})
	consumer := NewConsumer(funcName, ns, clientset, ready)
	errchan := group.Errors()

	<-ready
	for {
		select {
		case <-stopchan:
			return
		case err := <-errchan:
			logrus.Errorf("Kafka consumer brokers = %v topic = %v function = %v consumerID = %v: %v", brokers, topic, funcName, consumerGroupID, err)
		default:
		}

		if err := group.Consume(context.Background(), []string{topic}, consumer); err != nil {
			logrus.Errorf("Kafka consumer brokers = %v topic = %v function = %v consumerID = %v: %v", brokers, topic, funcName, consumerGroupID, err)
		}
	}
}

// CreateKafkaConsumer creates a goroutine that subscribes to Kafka topic
func CreateKafkaConsumer(triggerObjName, funcName, ns, topic string, clientset kubernetes.Interface) error {
	consumerID := generateUniqueConsumerGroupID(triggerObjName, funcName, ns, topic)
	if consumerM[consumerID] {
		logrus.Infof("Consumer for function %s associated with trigger %s already exists, so just returning", funcName, triggerObjName)
		return nil
	}

	logrus.Infof("Creating Kafka consumer for the function %s associated with for trigger %s", funcName, triggerObjName)
	stopM[consumerID] = make(chan struct{})
	stoppedM[consumerID] = make(chan struct{})
	go createConsumerProcess(topic, funcName, ns, consumerID, clientset, stopM[consumerID], stoppedM[consumerID])
	consumerM[consumerID] = true
	logrus.Infof("Created Kafka consumer for the function %s associated with for trigger %s", funcName, triggerObjName)

	return nil
}

// DeleteKafkaConsumer deletes goroutine created by CreateKafkaConsumer
func DeleteKafkaConsumer(triggerObjName, funcName, ns, topic string) error {
	consumerID := generateUniqueConsumerGroupID(triggerObjName, funcName, ns, topic)
	if !consumerM[consumerID] {
		logrus.Infof("Consumer for function %s associated with trigger %s doesn't exists. Good enough to skip the stop", funcName, triggerObjName)
		return nil
	}

	logrus.Infof("Stopping consumer for the function %s associated with for trigger %s", funcName, triggerObjName)
	// delete consumer process
	close(stopM[consumerID])
	<-stoppedM[consumerID]
	consumerM[consumerID] = false
	logrus.Infof("Stopped consumer for the function %s associated with for trigger %s", funcName, triggerObjName)

	return nil
}

func generateUniqueConsumerGroupID(triggerObjName, funcName, ns, topic string) string {
	return ns + "_" + triggerObjName + "_" + funcName + "_" + topic
}

// Consumer represents a Sarama consumer group consumer.
type Consumer struct {
	funcName, ns string
	clientset    kubernetes.Interface
	ready        chan struct{}
}

// NewConsumer returns new consumer.
func NewConsumer(funcName, ns string, clientset kubernetes.Interface, ready chan struct{}) *Consumer {
	return &Consumer{
		clientset: clientset,
		funcName:  funcName,
		ns:        ns,
		ready:     ready,
	}
}

// Setup is run at the beginning of a new session, before ConsumeClaim.
func (c *Consumer) Setup(sarama.ConsumerGroupSession) error {
	// Mark the consumer as ready
	close(c.ready)
	return nil
}

// Cleanup is run at the end of a session, once all ConsumeClaim goroutines have exited.
func (c *Consumer) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

// ConsumeClaim must start a consumer loop of ConsumerGroupClaim's Messages().
func (c *Consumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		logrus.Infof("Message claimed: value = %s, timestamp = %v, topic = %s", string(msg.Value), msg.Timestamp, msg.Topic)

		req, err := utils.GetHTTPReq(c.clientset, c.funcName, msg.Topic, c.ns, "kafkatriggers.kubeless.io", "POST", string(msg.Value))
		if err != nil {
			logrus.Errorf("Unable to elaborate request topic = %v function = %v: %v", msg.Topic, c.funcName, err)
			continue
		}

		if err = utils.SendMessage(req); err != nil {
			logrus.Errorf("Failed to send message topic = %v function = %v partition = %v offset = %v: %v", msg.Topic, c.funcName, msg.Partition, msg.Offset, err)
		} else {
			logrus.Infof("Message was sent to function successfully topic = %v function = %v partition = %v offset = %v", msg.Topic, c.funcName, msg.Partition, msg.Offset)
		}

		session.MarkMessage(msg, "")
	}
	return nil
}
