package rabbitmq

import (
	"github.com/Sirupsen/logrus"
	"github.com/streadway/amqp"
	"os"
	"sort"
)

var (
	log *logrus.Logger
	influxPointStoredChannel *amqp.Channel
	influxPointStoredQueue amqp.Queue
	rabbitMqConnection amqp.Connection
)


func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

func SetLogger(l *logrus.Logger) {
	log = l
}


func Init() {
	if rabbitMqConnection.IsClosed() || influxPointStoredChannel == nil {
		initConnection()
	}
}

func initConnection() {
	if os.Getenv("RABBITMQ_ENABLED") != "true" {
		return
	}

	rabbitmqUrl := os.Getenv("RABBITMQ_URL")
	if rabbitmqUrl == "" {
		rabbitmqUrl = "amqp://rabbitmq:rabbitmq@rabbitmq/" // default
	}

	if rabbitMqConnection.IsClosed() {
		defer rabbitMqConnection.Close()
	}

	rabbitMqConnection, err := amqp.Dial(rabbitmqUrl)
	failOnError(err, "Failed to connect to RabbitMQ")
	// defer rabbitMqConnection.Close()


	influxPointStoredChannel, err = rabbitMqConnection.Channel()
	failOnError(err, "Failed to open a channel")
	// defer influxPointStoredChannel.Close()


	influxPointStoredQueue, err = influxPointStoredChannel.QueueDeclare(
		"snmp_collector.influxpoint.stored", // name
		false,
		false,
		false,
		false,
		nil,
	)
	failOnError(err, "Failed to declare a queue")

}


func GetInfluxPointStored(measurement string, tags map[string]string) string {
	if os.Getenv("RABBITMQ_ENABLED") != "true" {
		return ""
	}

	body := "{ \"measurement\": \"" + measurement + "\""

	sortedKeys := sortKeys(tags) // be sure that body is always the same, key order could be different in map


	body += ", \"tags\": ["
	i := 0
	for _, tagKey := range sortedKeys {
		if tagKey == "subIndexesTag" {
			continue // we don't ned sub index tags for identification
		}

		if i > 0 {
			body += ", "
		}
		body += "{ \""  + tagKey + "\": \"" + tags[tagKey] + "\" }"
		i++
	}

	body += " ]}"
	return body
}

func PublishInfluxPointStored(influxPointsStored map[string]int) {
	if os.Getenv("RABBITMQ_ENABLED") != "true" {
		return
	}

	msg := "{ \"stored\": ["
	i := 0
	for influxPoint := range influxPointsStored {
		if i > 0 {
			msg += ", "
		}
		msg += " " + influxPoint
		i++
	}
	msg += "]}"


	// https://dzone.com/articles/try-and-catch-in-golang
	Block{
		Try: func() {

			if influxPointStoredChannel == nil {
				initConnection()
			}

			err := influxPointStoredChannel.Publish(
				"snmp_collector.influxpoint",
				influxPointStoredQueue.Name,
				false,
				false,
				amqp.Publishing{
					ContentType: "text/plain",
					Body:        []byte(msg),
				})
			log.Printf(" [x] Sent %s", msg)

			if err != nil {
				Throw(err)
			}

		},
		Catch: func(e Exception) {
			closeConnection()

		},
	}.Do()

}


func closeConnection() {
	Block{
		Try: func() {
			if influxPointStoredChannel != nil {
				defer influxPointStoredChannel.Close()
			}
			defer rabbitMqConnection.Close()
		},
		Catch: func(e Exception) {

		},
		Finally: func() {
			influxPointStoredChannel = nil
		},
	}.Do()
}




func sortKeys(m map[string]string) ([]string) {
	keys := make([]string, len(m))
	i := 0
	for k := range m {
		keys[i] = k
		i++
	}
	sort.Strings(keys)
	return keys
}



type Block struct {
	Try     func()
	Catch   func(Exception)
	Finally func()
}

type Exception interface{}

func Throw(up Exception) {
	panic(up)
}

func (tcf Block) Do() {
	if tcf.Finally != nil {

		defer tcf.Finally()
	}
	if tcf.Catch != nil {
		defer func() {
			if r := recover(); r != nil {
				tcf.Catch(r)
			}
		}()
	}
	tcf.Try()
}












