package eventprocessor

type EventProcessor interface {
	Process(name string)
	Close()
}
