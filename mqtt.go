package camera

import (
	"context"
	"encoding/gob"
	"github.com/eclipse/paho.mqtt.golang"
	"github.com/garyburd/redigo/redis"
	"github.com/sirupsen/logrus"
	"strings"
	"sync"
	"time"
)

func init() {
	gob.Register(map[string]interface{}{})
	gob.Register([]interface{}{})
	gob.Register([]map[string]int{})

	gob.Register(map[string]uint8{})
	gob.Register(map[string]int8{})
	gob.Register(map[string]int16{})
	gob.Register(map[string]uint16{})
	gob.Register(map[string]int32{})
	gob.Register(map[string]uint32{})
	gob.Register(map[string]int64{})
	gob.Register(map[string]uint64{})
	gob.Register(map[string]float32{})
	gob.Register(map[string]float64{})
	gob.Register(map[string]string{})
	gob.Register(map[string]int{})
	gob.Register(map[string][]map[string]int{})
}

// Config holds the MQTT pub-sub backend configuration.
type Config struct {
	Server       string
	Username     string
	Password     string
	QOS          uint8
	CleanSession bool
	ClientID     string
	CACert       string
	TLSCert      string
	TLSKey       string
}

// KafkaBackend implements a MQTT pub-sub backend.
type Backend struct {
	sync.RWMutex
	wg   sync.WaitGroup
	opts *mqtt.ClientOptions
	conn mqtt.Client

	deviceChan chan DevicePayload
	nodeChan   chan NodePayload

	config      Config
	rxTopic     string // 设备上报数据报文
	txTopic     string // 云端回应/下发命令
	ackTopic    string // 设备执行命令回复报文
	deviceTopic string // 设备下行数据

	ctx       context.Context
	redisPool *redis.Pool
}

var pubSub *Backend

// NewBackend creates a new NewSubset.
func NewBackend(c Config) error {

	b := Backend{
		deviceChan: make(chan DevicePayload, bufferSize1024),

		config:      c,
		deviceTopic: "/x55P94801qK/+/get",
		rxTopic:     "/x55P94801qK/7923463163321710/update",
		//txTopic:       "/%s/%s/get",
		ackTopic: "/x55P94801qK/+/ack",

		ctx: context.Background(),
	}

	opts, err := b.newOptions()
	if err != nil {
		return err
	}

	opts.SetAutoReconnect(false)
	opts.SetOnConnectHandler(b.onConnected)
	opts.SetConnectionLostHandler(b.onConnectionLost)
	b.opts = opts

	b.connectLoop()
	go handlerCameraChan(&b)

	pubSub = &b
	return nil
}

func (b *Backend) newOptions() (*mqtt.ClientOptions, error) {
	opts := mqtt.NewClientOptions()
	opts.AddBroker(b.config.Server)
	opts.SetUsername(b.config.Username)
	opts.SetPassword(b.config.Password)
	opts.SetCleanSession(true)
	opts.SetMaxReconnectInterval(time.Second * 5)
	opts.SetPingTimeout(time.Second * 60)
	opts.SetMessageChannelDepth(1024)
	opts.SetWriteTimeout(time.Second * 5)
	opts.SetAutoReconnect(false)
	opts.SetClientID(time.Now().String())
	tlsConfig, err := NewTLSConfig(b.config.CACert, b.config.TLSCert, b.config.TLSKey)
	if err != nil {
		return nil, err
	}
	if tlsConfig != nil {
		opts.SetTLSConfig(tlsConfig)
	}
	return opts, nil
}

// 关闭
func (b *Backend) Close() error {
	b.conn.Disconnect(0)
	b.wg.Wait()
	close(b.deviceChan)
	return nil
}

// 发送消息
func (b *Backend) publish(topic string, v []byte) error {
	if token := b.conn.Publish(topic, b.config.QOS, false, v); token.Wait() && token.Error() != nil {
		return token.Error()
	}
	return nil
}

// camera数据通道
func (b *Backend) deviceHandler(c mqtt.Client, msg mqtt.Message) {
	b.wg.Add(1)
	defer b.wg.Done()

	logrus.Debugf("topic %s,%s", msg.Topic(), string(msg.Payload()))

	s := strings.Split(msg.Topic(), "/")
	if len(s) != 4 {
		logrus.Error("topic split error")
		return
	}
	did := s[2]

	select {

	case b.deviceChan <- DevicePayload{Did: did, Payload: msg.Payload()}:
		return
	default:
		logrus.WithFields(logrus.Fields{"topic": msg.Topic()}).Error("device chan is full.")
		return
	}
}

// 启动连接
func (b *Backend) onConnected(c mqtt.Client) {
	if token := b.conn.Subscribe(b.deviceTopic, b.config.QOS, b.deviceHandler); token.Wait() && token.Error() != nil {
		logrus.WithField("topic", b.deviceTopic).Errorf("subscribe rx error: %s", token.Error())
	}
}

func (b *Backend) onConnectionLost(c mqtt.Client, reason error) {
	logrus.WithFields(logrus.Fields{
		"IsConnection": b.conn.IsConnected(),
	}).Errorf("connection lost error: %v", reason)
	b.disconnect()
	b.connectLoop()
}

func (b *Backend) connect() error {
	b.Lock()
	defer b.Unlock()

	b.conn = mqtt.NewClient(b.opts)

	if token := b.conn.Connect(); token.Wait() && token.Error() != nil {
		return token.Error()
	}

	return nil
}

// connectLoop blocks until the client is connected
func (b *Backend) connectLoop() {
	for {
		if err := b.connect(); err != nil {
			logrus.WithError(err).Error("connection error,sleep 2 second")
			time.Sleep(time.Second * 2)
		} else {
			break
		}
	}
}

func (b *Backend) disconnect() error {
	b.Lock()
	defer b.Unlock()

	b.conn.Disconnect(250)
	return nil
}

