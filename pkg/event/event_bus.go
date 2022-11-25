package event

import (
	"github.com/asaskevich/EventBus"
	"sync"
)

var bus EventBus.Bus
var onceDo sync.Once

func Bus() EventBus.Bus {
	onceDo.Do(func() {
		bus = EventBus.New()
	})
	return bus
}
