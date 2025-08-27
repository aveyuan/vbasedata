package vbasedata

import (
	"net"
	"time"
	"github.com/aveyuan/vlogger"

	"github.com/segmentio/kafka-go"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/segmentio/kafka-go/sasl"
	"github.com/segmentio/kafka-go/sasl/plain"
	"github.com/segmentio/kafka-go/sasl/scram"
)

const (
	SHA512 = "SHA512"
	SHA256 = "SHA256"
	PLAIN  = "PLAIN"
)

type KafkaConfig struct {
	Name      string   `json:"name" yaml:"name"`           //别名，用来区分多个kafka客户端
	Brokers   []string `json:"brokers" yaml:"brokers"`     //数组地址
	Topic     string   `json:"topic" yaml:"topic"`         //主题
	GroupId   string   `json:"groupid" yaml:"groupid"`     //消费者组
	Mechanism string   `json:"mechanism" yaml:"mechanism"` //机密机制
	Username  string   `json:"username" yaml:"username"`   //账号
	Password  string   `json:"password" yaml:"password"`   //密码
}

// NewKafkaReader 实例化kafka消费者
func NewKafkaReader(c *KafkaConfig, logger *log.Helper) (*kafka.Reader, error) {
	s, err := Sasl(c)
	if err != nil {
		return nil, err
	}
	dialer := &kafka.Dialer{
		Timeout:       10 * time.Second,
		DualStack:     true,
		SASLMechanism: s,
	}
	return kafka.NewReader(kafka.ReaderConfig{
		Brokers:     c.Brokers,
		GroupID:     c.GroupId,
		Topic:       c.Topic,
		MinBytes:    10e3, // 10KB
		MaxBytes:    10e6, // 10MB
		Dialer:      dialer,
		Logger:      vlogger.NewKafkaInfoLog(logger),
		ErrorLogger: vlogger.NewKafkaErrorLog(logger),
	}), nil
}

// NewKafkaWriter 实例化kafka生产者
func NewKafkaWriter(c *KafkaConfig) (*kafka.Writer, error) {
	s, err := Sasl(c)
	if err != nil {
		return nil, err
	}
	sharedTransport := &kafka.Transport{
		SASL: s,
	}
	w := kafka.Writer{
		Addr:      kafka.TCP(c.Brokers...),
		Topic:     c.Topic,
		Balancer:  &kafka.Hash{},
		Transport: sharedTransport,
	}
	w.AllowAutoTopicCreation = true
	return &w, nil
}

// Sasl kafka机密机制
func Sasl(c *KafkaConfig) (sasl.Mechanism, error) {
	// 判断kafka端口是否通畅
	for _, v := range c.Brokers {
		_, err := net.Dial("tcp", v)
		if err != nil {
			return nil, err
		}
	}
	// 判断账户密码
	var mechanism sasl.Mechanism
	var err error
	if c.Mechanism == PLAIN {
		mechanism = plain.Mechanism{
			Username: c.Username,
			Password: c.Password,
		}
	}

	if c.Mechanism == SHA256 {
		mechanism, err = scram.Mechanism(scram.SHA256, c.Username, c.Password)
		if err != nil {
			return nil, err
		}
	}

	if c.Mechanism == SHA512 {
		mechanism, err = scram.Mechanism(scram.SHA512, c.Username, c.Password)
		if err != nil {
			return nil, err
		}
	}
	return mechanism, nil
}
