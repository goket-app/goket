package eventprocessor

// Event specifies a single event in a configuration
type Event struct {
	// Type of action to run.
	Type string `json:"type"`
	// Action to run.
	Action string `json:"action"`
	// Timeout in seconds
	Timeout float64 `json:"timeout"`
	// Whether the "current" state should stay at this node when a child action was executed ; only until timeout occurs
	Stay bool `json:"stay"`
	// Map of child events
	Children EventMap
	parent   *Event
}

type EventMap map[string]*Event

// Initialize configures the EventMap after it was deserialized.
func (m *EventMap) Initialize() {
	m.setParent(nil)
}

// recursively set parent event for all children
func (m *EventMap) setParent(parent *Event) {
	for _, event := range *m {
		event.parent = parent
		if len(event.Children) > 0 {
			event.Children.setParent(event)
		}
	}
}
