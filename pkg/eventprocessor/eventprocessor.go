package eventprocessor

import "time"

type EventProcessor interface {
	Start()
	Close()
	Process(name string, when time.Time)
}
