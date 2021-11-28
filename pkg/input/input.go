package input

import "time"

type KeyboardInputEvent struct {
	Time    time.Time
	Down    bool
	KeyName string
}

type KeyboardInput interface {
	Read() (KeyboardInputEvent, error)
	Close() error
}
