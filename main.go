package main

import (
	"fmt"
	"log"
	"math/rand"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/IBM/sarama"
)

var (
	// Update with your Kafka broker addresses
	kafkaBrokers  = []string{"localhost:9092"} 
	// Update with your Kafka topic name
	topic = "test_topic" 
	// Set an appropriate buffer size based on your requirements             
	maxBufferSize = 100                       
)

// Global counter to keep track of the messages
var counter int 
// Internal slice to store messages received from Kafka
var messages = make([]string, 0, maxBufferSize)
// Mutex to synchronize access to the messages slice
var mutex = sync.Mutex{}
// Channel to send messages from the /producer endpoint to the Kafka producer goroutine
var producerMessages = make(chan string)
 // Channel to send messages from the Kafka consumer to the /consumer endpoint
var consumerMessages = make(chan string)

func kafkaProducer() {

	// Create a new Sarama configuration for the Kafka producer.
	config := sarama.NewConfig()

	// Create a new Kafka producer using the specified configuration and broker addresses.
	producer, err := sarama.NewAsyncProducer(kafkaBrokers, config)
	if err != nil {
		log.Fatal("Failed to start Kafka producer:", err)
	}

	// Ensure the Kafka producer is closed when the function ends (deferred execution).
	defer producer.Close()

	for message := range producerMessages {
		counterStr := fmt.Sprintf("%d", counter)

		// Get the Indian Standard Time (IST) location
		istLocation, err := time.LoadLocation("Asia/Kolkata")
		if err != nil {
			log.Fatal("Failed to load IST location:", err)
		}
		
		// Convert current time to IST
		istTime := time.Now().In(istLocation).Format("02-01-2006 15:04:05")
		value := fmt.Sprintf("(%s, %s, %s)", counterStr, istTime, message)
		producer.Input() <- &sarama.ProducerMessage{
			Topic: topic,
			Value: sarama.StringEncoder(value),
		}

		fmt.Printf("Message sent: %s\n", value)
		counter++

		// Introduce random delay between 1 to 3 seconds for message push
		time.Sleep(time.Duration(1+rand.Intn(3)) * time.Second)
	}
}

func kafkaConsumer(wg *sync.WaitGroup) {

	// Create a new Sarama configuration for the Kafka producer.
	config := sarama.NewConfig()
	
	// Create a new Kafka consumer using the specified configuration and broker addresses.
	consumer, err := sarama.NewConsumer(kafkaBrokers, config)
	if err != nil {
		log.Fatal("Failed to start Kafka consumer:", err)
	}

	// Ensure the Kafka consumer is closed when the function ends (deferred execution).
	defer consumer.Close()

	// Create a partition consumer for the specified topic, partition, and starting offset.
	// The starting offset is set to sarama.OffsetNewest, which means the consumer will start consuming messages from the latest available offset.
	partitionConsumer, err := consumer.ConsumePartition(topic, 0, sarama.OffsetNewest) 
	if err != nil {
		log.Fatal("Failed to start partition consumer:", err)
	}

	// Ensure the partition consumer is closed when the function ends (deferred execution).
	defer partitionConsumer.Close()

	// Signal that the consumer goroutine is ready
	wg.Done()

	// Infinite loop to continuously listen for messages from the partitionConsumer.Messages() channel.
	for {
		select {
		case message := <-partitionConsumer.Messages():
			value := string(message.Value)
			fmt.Printf("Received message from Kafka: %s\n", value)
			// Acquire the mutex before appending to the messages slice to ensure concurrent-safe access.
			mutex.Lock()
			// Append the received message to the internal messages slice.
			messages = append(messages, value)
			// Release the mutex.
			mutex.Unlock()
			 // Send the received message to the /consumer endpoint via the consumerMessages channel.
			consumerMessages <- value
		}
	}
}

func main() {
	// Create a new instance of the Fiber web framework.	
	app := fiber.New()

	// Create a WaitGroup to synchronize goroutines.
	wg := &sync.WaitGroup{}
	// Add 1 to the WaitGroup, indicating one goroutine to wait for.
	wg.Add(1)

	// Launch the Kafka producer goroutine in the background.
	go kafkaProducer()

	// Launch the Kafka consumer goroutine in the background, passing the WaitGroup for synchronization.
	go kafkaConsumer(wg)

	// Wait for the consumer goroutine to be ready
	wg.Wait()

	// The /producer endpoint for sending messages to the Kafka producer
	app.Get("/producer/:message", func(c *fiber.Ctx) error {
		message := c.Params("message")
		// Sending message to the Kafka producer via the producerMessages channel
		producerMessages <- message 
		return c.SendString("Message sent to Kafka producer.")
	})

	// The /consumer endpoint for receiving messages from the Kafka consumer
	app.Get("/consumer", func(c *fiber.Ctx) error {
		select {
		case msg := <-consumerMessages:
			 // If a message is available in the consumerMessages channel, return it as the response.
			return c.SendString(fmt.Sprintf("Received from Kafka consumer: %s", msg))
		case <-time.After(4 * time.Second): 
			// If no message is available within 4 seconds, respond with a default message.
			return c.SendString("No messages available at the moment. Please try again later.")
		}
	})

	log.Fatal(app.Listen(":3000"))
}
