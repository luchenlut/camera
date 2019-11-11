package camera

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/olivere/elastic"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"sync"
	"time"
)

const (
	Type = "log"
)

//elastic search
type ESClient struct {
	Index string
	Es    *elastic.Client
}

func NewESClients(es *ESConfig) *ESClient {
	if !es.Close {
		ctx := context.Background()
		var err error
		esClient, err := elastic.NewClient(
			elastic.SetURL(es.Url),
			elastic.SetScheme(es.Scheme),
			elastic.SetHealthcheck(true), //true时, 设置健康检查
			elastic.SetHealthcheckInterval(10*time.Second),
			elastic.SetHealthcheckTimeout(1*time.Second),
			elastic.SetHealthcheckTimeoutStartup(2*time.Second),
			elastic.SetSniff(false), //true的时候,设置监测interval:SetSnifferInterval,SetSnifferTimeoutStartup,SetSnifferTimeout
			elastic.SetSendGetBodyAs("GET"),
			elastic.SetBasicAuth(es.Username, es.Password),
		)
		if err != nil {
			return nil
		}
		_, _, err = esClient.Ping(es.Url).Do(ctx)
		if err != nil {
			logrus.Fatalf("ping es error %v", err)
			return nil
		}
		return &ESClient{es.Index, esClient}
	}
	return &ESClient{}
}

func (es *ESClient) write(event *event) error {
	if es.Es == nil {
		return errors.New("es client not connection")
	}

	if _, err := es.Es.Index().Index(event.Index).Type(Type).BodyJson(event).Do(context.Background()); err != nil {
		return err
	}

	return nil
}

type Writer interface {
	write(*event) error
}

type ConfigI interface {
	Init() Writer
}

type MonGoConfig struct {
	Url string
}

type ESConfig struct {
	Url      string
	Index    string
	Scheme   string
	Username string
	Password string
	Close    bool
}

func (EC *ESConfig) Init() Writer {
	if EC.Index == "" {
		EC.Index = "iot"
	}
	if EC.Scheme == "" {
		EC.Scheme = "http"
	}
	if EC.Url == "" {
		EC.Url = "http://192.168.1.6:9200"
	}
	return NewESClients(EC)
}

const DeviceLogIndex = "device"

type Fields map[string]interface{}

// The FieldLogger interface generalizes the Entry and Logger types
type FieldLogger interface {
	WithFields(fields Fields) *Entry

	Error(format string, args ...interface{})
	Active(format string, args ...interface{})
	UpLink(format string, args ...interface{})
	DownLink(format string, args ...interface{})
	UpgradeUp(format string, args ...interface{})
	UpgradeDown(format string, args ...interface{})
}

func NewEntry(fields Fields) *Entry {
	entry := new(Entry)
	entry.WithFields(fields)
	return entry
}

const (
	// 错误日志
	ErrorLevel Level = iota + 1
	// 激活日志
	ActiveLevel
	// 上行日志
	UpLinkLevel
	// 下行日志
	DownLinkLevel
	// 上线日志
	OnLineLevel
	// 下线日志
	OffLineLevel
	// 升级上行日志
	UpgradeUpLevel
	// 升级下行日志
	UpgradeDownLevel
	// 设备攻击
	AttackLevel
)

// Level type
type Level uint32

// Convert the Level to a string. E.g. PanicLevel becomes "panic".
func (level Level) String() string {
	switch level {
	case ErrorLevel:
		return "error"
	case ActiveLevel:
		return "active"
	case UpLinkLevel:
		return "upLink"
	case DownLinkLevel:
		return "downLink"
	case OnLineLevel:
		return "online"
	case OffLineLevel:
		return "offline"
	case UpgradeUpLevel:
		return "upgradeUp"
	case AttackLevel:
		return "Attack"
	case UpgradeDownLevel:
		return "upgradeDown"
	}
	return "unknown"
}

type Entry struct {
	Tid     int32
	Pid     int32
	Did     string
	Data    Fields
	Level   Level
	Message string
}

func (entry *Entry) WithFields(fields Fields) *Entry {
	data := make(Fields, len(entry.Data)+len(fields))
	for k, v := range entry.Data {
		data[k] = v
	}
	for k, v := range fields {
		switch k {
		case "tid":
			entry.Tid = v.(int32)
		case "pid":
			entry.Pid = v.(int32)
		case "did":
			entry.Did = v.(string)
		default:
			data[k] = v
		}
	}
	entry.Data = data
	return entry
}

func (entry *Entry) Error(format string, args ...interface{}) {
	entry.Level = ErrorLevel
	entry.echo(format, args...)
}

func (entry *Entry) Active(format string, args ...interface{}) {
	entry.Level = ActiveLevel
	entry.echo(format, args...)
}

func (entry *Entry) UpLink(format string, args ...interface{}) {
	entry.Level = UpLinkLevel
	entry.echo(format, args...)
}

func (entry *Entry) DownLink(format string, args ...interface{}) {
	entry.Level = DownLinkLevel
	entry.echo(format, args...)
}

func (entry *Entry) UpgradeUp(format string, args ...interface{}) {
	entry.Level = UpgradeUpLevel
	entry.echo(format, args...)
}

func (entry *Entry) UpgradeDown(format string, args ...interface{}) {
	entry.Level = UpgradeDownLevel
	entry.echo(format, args...)
}

func (entry *Entry) Attack(format string, args ...interface{}) {
	entry.Level = AttackLevel
	entry.echo(format, args...)
}

func (entry *Entry) echo(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	if len(entry.Data) > 0 {
		bytes, _ := json.Marshal(entry.Data)
		entry.Message = message + "\t\t" + string(bytes)
	} else {
		entry.Message = message
	}

	if entry.Level == ErrorLevel {
		entry.Message = `<span style="color:red;">` + entry.Message + `</span>`
	}

	if err := Mark(MarkFields{
		"tid":     entry.Tid,
		"pid":     entry.Pid,
		"did":     entry.Did,
		"type":    entry.Level,
		"message": entry.Message,
	}, DeviceLogIndex); err != nil {
		logrus.Error(err)
	}

	logrus.WithFields(logrus.Fields{
		"tid": entry.Tid,
		"pid": entry.Pid,
		"did": entry.Did,
	}).Debug(entry.Message)
}

const (
	eventBufferSize = 1000 * 1024
)

type MarkFields map[string]interface{}

type event struct {
	Index string                 `json:"index"`
	Value map[string]interface{} `json:"value"`
}

type reporter struct {
	stopping int32
	eventBus chan *event
	interval time.Duration
	writer   Writer
	evtBuf   *sync.Pool
}

var (
	reg = &reporter{
		stopping: 0,
		eventBus: make(chan *event, eventBufferSize),
		evtBuf:   &sync.Pool{New: func() interface{} { return new(event) }},
	}
)

func Run(i ConfigI) error {
	reg.writer = i.Init()
	go reg.eventLoop()
	return nil
}

func (r *reporter) eventLoop() {
	for {
		select {
		case evt, ok := <-r.eventBus:
			if !ok {
				break
			} else {
				r.writer.write(evt)
			}
		}
	}
}

func Mark(field MarkFields, index string) error {
	evt := reg.evtBuf.Get().(*event)
	evt.Value = field
	evt.Value["timestamp"] = time.Now()
	evt.Index = index
	select {
	case reg.eventBus <- evt:
	default:
		return errors.New("metrics eventBus is full.")
	}
	return nil
}
