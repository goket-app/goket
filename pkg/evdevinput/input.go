package evdevinput

import (
	"time"

	"github.com/goket-app/goket/pkg/input"
	evdev "github.com/holoplot/go-evdev"
)

type evdevKeyboardInput struct {
	input *evdev.InputDevice
}

func NewEvdevKeyboardInput(device string) (input.KeyboardInput, error) {
	input, err := evdev.Open(device)
	if err != nil {
		return nil, err
	}
	i := &evdevKeyboardInput{
		input: input,
	}
	return i, nil
}

func (i *evdevKeyboardInput) Read() (input.KeyboardInputEvent, error) {
	for {
		e, err := i.input.ReadOne()
		if err != nil {
			return input.KeyboardInputEvent{}, err
		}

		if e.Type == evdev.EV_KEY {
			keyName, ok := evdev.KEYName[e.Code]
			if ok {
				return input.KeyboardInputEvent{
					Time:    time.Unix(int64(e.Time.Sec), int64(e.Time.Usec)*1000),
					Down:    e.Value != 0,
					KeyName: keyName,
				}, nil
			}
		}
	}
}

func (i *evdevKeyboardInput) Close() error {
	i.input.Close()

	return nil
}
