package eventbus

import (
	"fmt"

	"github.com/gusti-andika/domino/ui"
)

type EventType int32

const (
	PlayerJoined EventType = iota
	CardAssigned
	InitialCard
	SkipTurn
	InvalidMove
	PlayerMoved
	GameFinished
	GameLog
)

type Event struct {
	Type EventType
	Data map[string]interface{}
}

type eventBus struct {
	ch      chan Event
	handler map[EventType]func(Event)
}

var (
	bus eventBus = eventBus{
		ch:      make(chan Event, 10),
		handler: map[EventType]func(Event){},
	}
	log     *ui.LogWindow
	started bool
)

func SetLog(t *ui.LogWindow) {
	log = t
}

func Start() {

	go func() {
		for event := range bus.ch {
			write(fmt.Sprintf("Process Event: %v", event))

			if handler, ok := bus.handler[event.Type]; ok {
				handler(event)
			}
		}
	}()
	started = true
	write(fmt.Sprintf("Eventbus started:%+v", bus))
}

func write(s string) {
	if log != nil {
		log.Log(s + "\n")
	}
}

func Post(event Event) {
	write(fmt.Sprintf("--->Post: %v", event))
	if !started {
		return
	}
	//if bus != nil {
	//go func() {
	//	write(fmt.Sprintf("(2)Received Event: %v", event))
	bus.ch <- event
	//}
	//}
	write(fmt.Sprintf("<---Post: %v", event))

}

func Close() {
	close(bus.ch)
}

func AddHandler(t EventType, handler func(Event)) {
	//if bus != nil {
	bus.handler[t] = handler
	//}
}
