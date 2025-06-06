package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strconv"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

const kafkaServerKey = "KAFKA_SERVER"
const kafkaPortNumberKey = "KAFKA_PORT_NUM"
const kafkaTopicKey = "KAFKA_TOPIC"

type Message struct {
	Type   int `json:"type"`
	Number int `json:"number"`
}

func main() {

	// get configuration
	kafkaServer, kafkaPort, kafkaTopic := getConfig()

	// Create Kafka producer
	producer, err := kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers": fmt.Sprintf("%s:%d", kafkaServer, kafkaPort),
	})
	if err != nil {
		log.Fatalf("Failed to create producer: %s", err)
	}
	defer producer.Close()

	// Create a channel to signal when a key is pressed, launch go routine
	stopChan := make(chan bool)
	go mainLoop(kafkaTopic, producer, stopChan)

	// Wait for a key press
	fmt.Println("Press any key to stop the loop...")
	reader := bufio.NewReader(os.Stdin)
	_, _ = reader.ReadByte() // Read one byte (any key)

	// Signal the goroutine to stop
	stopChan <- true

	// Give some time for the goroutine to clean up
	time.Sleep(100 * time.Millisecond)
	fmt.Println("Loop stopped. Exiting program.")
}

func mainLoop(kafkaTopic string, producer *kafka.Producer, stopChan chan bool) {

	// Infinite loop to generate and send messages
	for {
		select {
		case <-stopChan:
			fmt.Println("Stopping mainLoop...")
			return
		default:

			// Generate random number
			randomNumber := rand.Intn(1000000) + 1

			// Create message
			msg := Message{
				Type:   1,
				Number: randomNumber,
			}

			// Serialize message to JSON
			msgBytes, err := json.Marshal(msg)
			if err != nil {
				log.Printf("Failed to serialize message: %s", err)
				continue
			}

			// Send message to Kafka
			err = producer.Produce(&kafka.Message{
				TopicPartition: kafka.TopicPartition{Topic: &kafkaTopic,
					Partition: kafka.PartitionAny},
				Value: msgBytes,
			}, nil)

			if err != nil {
				log.Printf("Failed to produce message: %s", err)
			} else {
				fmt.Printf("Produced message: %s\n", msgBytes)
			}

			// Wait for 1 second before sending the next message
			time.Sleep(1 * time.Second)
		}
	}
}

func getConfig() (string, int, string) {
	kafkaServerString := os.Getenv(kafkaServerKey)
	kafkaPortNumberString := os.Getenv(kafkaPortNumberKey)
	topicString := os.Getenv(kafkaTopicKey)

	if (kafkaServerString == "") ||
		(kafkaPortNumberString == "") ||
		(topicString == "") {
		usage()
	}

	kafkaPortNumber, err := strconv.Atoi(kafkaPortNumberString)
	if err != nil {
		usage()
	}

	return kafkaServerString, kafkaPortNumber, topicString
}

func usage() {
	fmt.Printf("\nUSAGE: RandomNumberService\n\nThe following environment vars must be set:\n%s, %s, %s\n\n",
		kafkaServerKey, kafkaPortNumberKey, kafkaTopicKey)
	os.Exit(1)
}
