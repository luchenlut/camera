package echo

import (
	"github.com/pkg/errors"
	"sync"
	"time"
)

const (
	eventBufferSize = 1000 * 1024
)

type MarkFields map[string]interface{}

type event struct {
	Index string
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
	sinkDuration = time.Second * 5
	reg          = &reporter{
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
