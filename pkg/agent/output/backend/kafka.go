package backend

// The MIT License (MIT)
// Copyright (c) 2015-2020 InfluxData Inc.

import (
	_ "embed"
	"fmt"
	"strings"
	"time"

	"github.com/Shopify/sarama"
	"github.com/gofrs/uuid"
	"github.com/sirupsen/logrus"
	"github.com/toni-moreno/snmpcollector/pkg/config"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/common/kafka"
	"github.com/influxdata/telegraf/plugins/common/proxy"
	"github.com/influxdata/telegraf/plugins/common/tls"
	"github.com/influxdata/telegraf/plugins/serializers"
)

var sampleConfig string

var ValidTopicSuffixMethods = []string{
	"",
	"measurement",
	"tags",
}

var zeroTime = time.Unix(0, 0)

func NewNotInitKafka(c *config.KafkaCfg) (*Kafka, error) {

	sc := &serializers.Config{TimestampUnits: 1 * time.Second}
	sc.DataFormat = "json"
	ssc, err := serializers.NewSerializer(sc)
	if err != nil {
		log.Errorf("unable to set serializer: %s", err)
		return nil, err
	}

	// build own telegraf kafka config from plain struct
	tlsconf := tls.ClientConfig{
		TLSCA:              c.TLSCA,
		TLSCert:            c.TLSCert,
		TLSKey:             c.TLSKey,
		TLSKeyPwd:          c.TLSKeyPwd,
		TLSMinVersion:      c.TLSMinVersion,
		InsecureSkipVerify: c.InsecureSkipVerify,
		ServerName:         c.ServerName,
	}
	conf := kafka.Config{
		ClientConfig:     tlsconf,
		Version:          c.Version,
		CompressionCodec: c.CompressionCodec,
		EnableTLS:        &c.EnableTLS,
		MetadataFull:     &c.MetadataFull,
	}

	tsuffix := TopicSuffix{
		Method:    c.Method,
		Keys:      c.Keys,
		Separator: c.Separator,
	}

	kafkaserver := &Kafka{
		Brokers:         c.Brokers,
		Topic:           c.Topic,
		ExcludeTopicTag: c.ExcludeTopicTag,
		TopicTag:        c.TopicTag,
		RoutingTag:      c.RoutingTag,
		RoutingKey:      c.RoutingKey,
		TopicSuffix:     tsuffix,
		WriteConfig: kafka.WriteConfig{
			Config:       conf,
			MaxRetry:     c.MaxRetry,
			RequiredAcks: c.RequiredAcks,
		},
		producerFunc: sarama.NewSyncProducer,
		Log:          *log,
	}
	kafkaserver.SetSerializer(ssc)
	return kafkaserver, nil
}

type Kafka struct {
	Brokers         []string
	Topic           string
	TopicTag        string
	ExcludeTopicTag bool
	TopicSuffix     TopicSuffix
	RoutingTag      string
	RoutingKey      string

	proxy.Socks5ProxyConfig

	// Legacy TLS config options
	// TLS client certificate
	Certificate string
	// TLS client key
	Key string
	// TLS certificate authority
	CA string

	kafka.WriteConfig

	Log logrus.Logger

	saramaConfig *sarama.Config
	producerFunc func(addrs []string, config *sarama.Config) (sarama.SyncProducer, error)
	producer     sarama.SyncProducer

	serializer serializers.Serializer
}

type TopicSuffix struct {
	Method    string
	Keys      []string
	Separator string
}

// DebugLogger logs messages from sarama at the debug level.
type DebugLogger struct {
	Log *logrus.Logger
}

func (l *DebugLogger) Print(v ...interface{}) {
	args := make([]interface{}, 0, len(v)+1)
	args = append(append(args, "[sarama] "), v...)
	log.Debug(args...)
}

func (l *DebugLogger) Printf(format string, v ...interface{}) {
	log.Debugf("[sarama] "+format, v...)
}

func (l *DebugLogger) Println(v ...interface{}) {
	l.Print(v)
}

func ValidateTopicSuffixMethod(method string) error {
	for _, validMethod := range ValidTopicSuffixMethods {
		if method == validMethod {
			return nil
		}
	}
	return fmt.Errorf("unknown topic suffix method provided: %s", method)
}

func (*Kafka) SampleConfig() string {
	return sampleConfig
}

func (k *Kafka) GetTopicName(metric telegraf.Metric) (telegraf.Metric, string) {
	topic := k.Topic
	if k.TopicTag != "" {
		if t, ok := metric.GetTag(k.TopicTag); ok {
			topic = t

			// If excluding the topic tag, a copy is required to avoid modifying
			// the metric buffer.
			if k.ExcludeTopicTag {
				metric = metric.Copy()
				metric.Accept()
				metric.RemoveTag(k.TopicTag)
			}
		}
	}

	var topicName string
	switch k.TopicSuffix.Method {
	case "measurement":
		topicName = topic + k.TopicSuffix.Separator + metric.Name()
	case "tags":
		var topicNameComponents []string
		topicNameComponents = append(topicNameComponents, topic)
		for _, tag := range k.TopicSuffix.Keys {
			tagValue := metric.Tags()[tag]
			if tagValue != "" {
				topicNameComponents = append(topicNameComponents, tagValue)
			}
		}
		topicName = strings.Join(topicNameComponents, k.TopicSuffix.Separator)
	default:
		topicName = topic
	}
	return metric, topicName
}

func (k *Kafka) SetSerializer(serializer serializers.Serializer) {
	k.serializer = serializer
}

func (k *Kafka) Init() error {
	sarama.Logger = &DebugLogger{Log: log}

	err := ValidateTopicSuffixMethod(k.TopicSuffix.Method)
	if err != nil {
		return err
	}
	config := sarama.NewConfig()

	if err := k.SetConfig(config); err != nil {
		return err
	}

	k.saramaConfig = config

	// Legacy support ssl config
	if k.Certificate != "" {
		k.TLSCert = k.Certificate
		k.TLSCA = k.CA
		k.TLSKey = k.Key
	}

	if k.Socks5ProxyEnabled {
		config.Net.Proxy.Enable = true

		dialer, err := k.Socks5ProxyConfig.GetDialer()
		if err != nil {
			return fmt.Errorf("connecting to proxy server failed: %s", err)
		}
		config.Net.Proxy.Dialer = dialer
	}

	return nil
}

func (k *Kafka) Connect() error {
	k.Init()
	producer, err := k.producerFunc(k.Brokers, k.saramaConfig)
	if err != nil {
		return err
	}
	k.producer = producer
	return nil
}

func (k *Kafka) Close() error {
	return k.producer.Close()
}

func (k *Kafka) routingKey(metric telegraf.Metric) (string, error) {
	if k.RoutingTag != "" {
		key, ok := metric.GetTag(k.RoutingTag)
		if ok {
			return key, nil
		}
	}

	if k.RoutingKey == "random" {
		u, err := uuid.NewV4()
		if err != nil {
			return "", err
		}
		return u.String(), nil
	}

	return k.RoutingKey, nil
}

func (k *Kafka) Write(metrics []telegraf.Metric) error {
	msgs := make([]*sarama.ProducerMessage, 0, len(metrics))
	for _, metric := range metrics {
		metric, topic := k.GetTopicName(metric)

		buf, err := k.serializer.Serialize(metric)
		if err != nil {
			log.Debugf("Could not serialize metric: %v", err)
			continue
		}

		m := &sarama.ProducerMessage{
			Topic: topic,
			Value: sarama.ByteEncoder(buf),
		}

		// Negative timestamps are not allowed by the Kafka protocol.
		if !metric.Time().Before(zeroTime) {
			m.Timestamp = metric.Time()
		}

		key, err := k.routingKey(metric)
		if err != nil {
			return fmt.Errorf("could not generate routing key: %v", err)
		}

		if key != "" {
			m.Key = sarama.StringEncoder(key)
		}
		msgs = append(msgs, m)
	}

	err := k.producer.SendMessages(msgs)
	if err != nil {
		// We could have many errors, return only the first encountered.
		if errs, ok := err.(sarama.ProducerErrors); ok {
			for _, prodErr := range errs {
				if prodErr.Err == sarama.ErrMessageSizeTooLarge {
					k.Log.Error("Message too large, consider increasing `max_message_bytes`; dropping batch")
					return nil
				}
				if prodErr.Err == sarama.ErrInvalidTimestamp {
					k.Log.Error("The timestamp of the message is out of acceptable range, consider increasing broker `message.timestamp.difference.max.ms`; dropping batch")
					return nil
				}
				return prodErr //nolint:staticcheck // Return first error encountered
			}
		}
		return err
	}

	return nil
}
