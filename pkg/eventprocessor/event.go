package eventprocessor

type Event struct {
	Type     string  `json:"type"`
	Action   string  `json:"action"`
	Timeout  float64 `json:"timeout"`
	Children EventMap
}

type EventMap map[string]*Event
