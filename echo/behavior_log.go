package echo

import (
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
)

const DeviceLogIndex = "device"

type Fields map[string]interface{}

// The FieldLogger interface generalizes the Entry and Logger types
type FieldLogger interface {
	WithFields(fields Fields) *Entry
	WithError(err error) *Entry

	Error(format string, args ...interface{})
	Active(format string, args ...interface{})
	UpLink(format string, args ...interface{})
	DownLink(format string, args ...interface{})
	OnLine(format string, args ...interface{})
	OffLine(format string, args ...interface{})
	UpgradeUp(format string, args ...interface{})
	UpgradeDown(format string, args ...interface{})
	Attack(format string, args ...interface{})
}

//func Log(tid,pid int32,did string) *Entry {
//	return WithField(Fields{"tid":tid,"pid":pid,"did":did})
//}

func WithField(fields Fields) *Entry {
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

func (entry *Entry) WithError(err error) *Entry {
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

func (entry *Entry) OnLine(format string, args ...interface{}) {
	entry.Level = OnLineLevel
	entry.echo(format, args...)
}

func (entry *Entry) OffLine(format string, args ...interface{}) {
	entry.Level = OffLineLevel
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
		log.Error(err)
	}

	log.WithFields(log.Fields{
		"tid": entry.Tid,
		"pid": entry.Pid,
		"did": entry.Did,
	}).Debug(entry.Message)
}
