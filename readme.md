# Go Web Server with Kafka Producer and Consumer using Fiber

This is a simple Go web server that implements a Kafka producer and consumer using the Fiber web framework. The server has two endpoints: one for sending data to the Kafka producer and another for receiving data from the Kafka consumer.

## Prerequisites

Before running the application, you need to have the following installed:
- Golang (https://go.dev/dl/)
- Kafka (https://kafka.apache.org/downloads)
- Zookeeper (https://zookeeper.apache.org/releases.html)

## Setup

Clone the repository:

```
git clone https://github.com/Sahil-4555/go-kafka.git
```

Install the required dependencies:

```
go mod tidy
```

Make sure you have Kafka running on localhost:9092 or update the kafkaBrokers variable in main.go with your Kafka broker addresses.

## Start ZooKeeper

Apache Kafka relies on ZooKeeper for maintaining metadata. You need to start ZooKeeper before starting Kafka.

Open a terminal in the folder where Kafka is downloaded, and then run ZooKeeper using the following command:

```
.\bin\windows\zookeeper-server-start.bat .\config\zookeeper.properties
```

## Start Kafka Broker

Open another terminal in the folder where kafka is downloaded, and start the kafka server using the following command:

```
.\bin\windows\kafka-server-start.bat .\config\server.properties
```

If You want to consume and print messages from the Kafka topic in real-time, starting from the beginning of the topic, while connecting to the Kafka broker running on localhost:9092. run the following command in another terminal:

```
.\bin\windows\kafka-console-consumer.bat --topic <topic name> --bootstrap-server localhost:9092 --from-beginning
```

## Running the application

To run the application, execute the following command:

```
go run main.go
```

The server will start listening on `http://localhost:3000`.

## Endpoints

To send data to the Kafka producer, make a GET request to the /producer endpoint with a message as a parameter.

```
// GET METHOD
http://localhost:3000/producer/:message
```

To receive data from the Kafka consumer, make a GET request to the /consumer endpoint. The endpoint will respond with the last received message from the Kafka consumer or a default message if no message is available within 4 seconds.

```
// GET METHOD
http://localhost:3000/consumer
```

## How it Works

The application uses the Fiber web framework to handle HTTP requests. When a client sends a GET request to the `/producer` endpoint with a message, the message is sent to the Kafka producer via the `producerMessages` channel.

The Kafka producer, running in a separate goroutine, reads messages from the `producerMessages` channel and sends them to the Kafka topic. The producer introduces a random delay between 1 to 3 seconds for message push to simulate real-world scenarios.

The Kafka consumer, also running in a separate goroutine, continuously listens for new messages from the Kafka topic. When a new message is received, it is added to the internal `messages` slice using a mutex to ensure concurrent-safe access.

The `/consumer` endpoint, upon receiving a GET request, checks the `consumerMessages` channel for any available message within 4 seconds. If a message is available, it is sent as a response. Otherwise, a default message is sent indicating that no messages are available at the moment.

The global counter `counter` is used to keep track of the messages sent to the Kafka producer.







