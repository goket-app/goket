package eventprocessor

import (
	"testing"
	"time"

	"go.uber.org/zap/zaptest"
)

var (
	timeout         = 1.0                                                  // timeout in seconds
	timeoutDuration = time.Duration(int64(timeout * float64(time.Second))) // timeout as duration
	startTime       = time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	defaultEventMap = map[string]*Event{
		"a": {
			Action: "action://a",
		},
		"c": {
			Children: map[string]*Event{
				"a": {
					Action: "action://c-a",
				},
				"c": {
					Action: "action://c-c",
					Children: map[string]*Event{
						"a": {
							Action: "action://c-c-a",
						},
					},
				},
				"t": {
					// increase timeout, test custom timeout handling
					Timeout: 10 * timeout,
					Children: map[string]*Event{
						"a": {
							Action: "action://c-t-a",
						},
					},
				},
			},
		},
		"s": {
			Timeout: 10 * timeout,
			Stay:    true,
			Children: map[string]*Event{
				"t": {
					Timeout: 5 * timeout,
					Children: map[string]*Event{
						"a": {
							Action: "action://s-t-a",
						},
					},
				},
				"a": {
					Action: "action://s-a",
				},
			},
		},
	}
)

type testEvent struct {
	name  string
	after time.Duration
}

func setupProcessor(t *testing.T, eventMap EventMap) (*processor, chan *Event) {
	out := make(chan *Event, 256)
	logger := zaptest.NewLogger(t)
	p := NewProcessor(logger, eventMap, timeout, out)
	processor, ok := p.(*processor)
	if !ok {
		t.Fatalf("unable to cast %v as processor", p)
	}
	return processor, out
}

func TestEvents(t *testing.T) {
	for _, test := range []struct {
		name     string
		eventMap EventMap
		inputs   []*testEvent
		expected []string
	}{
		{
			name: "event with action",
			inputs: []*testEvent{
				{name: "a"},
			},
			expected: []string{
				"action://a",
			},
		},

		{
			name: "event with child only",
			inputs: []*testEvent{
				{name: "c"},
			},
			expected: []string{},
		},

		{
			name: "event with child action, within timeout",
			inputs: []*testEvent{
				{name: "c"},
				{name: "a", after: timeoutDuration / 2},
			},
			expected: []string{
				"action://c-a",
			},
		},

		{
			name: "event with child action, outside timeout",
			inputs: []*testEvent{
				{name: "c"},
				{name: "a", after: timeoutDuration * 2},
			},
			expected: []string{
				// expect "a" as the C -> A traversal timed out
				"action://a",
			},
		},

		{
			name: "event with children and action",
			inputs: []*testEvent{
				{name: "c"},
				{name: "c", after: timeoutDuration / 2},
			},
			expected: []string{
				"action://c-c",
			},
		},

		{
			name: "call child from parent with action, within timeout",
			inputs: []*testEvent{
				{name: "c"},
				{name: "c", after: timeoutDuration / 2},
				{name: "a", after: timeoutDuration / 2},
			},
			expected: []string{
				"action://c-c-a",
			},
		},

		{
			name: "call child from parent with action, outside timeout",
			inputs: []*testEvent{
				{name: "c"},
				{name: "c", after: timeoutDuration * 2},
				{name: "a", after: timeoutDuration / 2},
			},
			expected: []string{
				// expect C->A as the C->C traversal timed out, hence second "c" was traversed from root
				"action://c-a",
			},
		},

		{
			name: "call parent with action, timing out on second keystroke",
			inputs: []*testEvent{
				{name: "c"},
				{name: "c", after: timeoutDuration / 2},
				{name: "a", after: timeoutDuration * 2},
			},
			expected: []string{
				"action://c-c",
				"action://a",
			},
		},

		{
			name: "call child with different timeout, within timeout",
			inputs: []*testEvent{
				{name: "c"},
				{name: "t", after: timeoutDuration / 2},
				{name: "a", after: 9.0 * time.Second},
			},
			expected: []string{
				"action://c-t-a",
			},
		},

		{
			name: "call child with different timeout, outside timeout",
			inputs: []*testEvent{
				{name: "c"},
				{name: "t", after: timeoutDuration / 2},
				{name: "a", after: 11.0 * time.Second},
			},
			expected: []string{
				"action://a",
			},
		},

		{
			name: "ensure timeout is reset after traversal through node with custom timeout, within timeout",
			inputs: []*testEvent{
				{name: "c"},
				{name: "t", after: timeoutDuration / 2},
				{name: "a", after: 9 * timeoutDuration},
				{name: "c", after: timeoutDuration / 2},
				{name: "a", after: timeoutDuration * 2},
			},
			expected: []string{
				"action://c-t-a",
				// expect A as the C->A traversal timed out (this is only true if timeout was reset)
				"action://a",
			},
		},

		{
			name: "ensure timeout is reset after traversal through node with custom timeout, outside timeout",
			inputs: []*testEvent{
				{name: "c"},
				{name: "t", after: timeoutDuration / 2},
				{name: "a", after: 11 * timeoutDuration},
				{name: "c", after: timeoutDuration / 2},
				{name: "a", after: timeoutDuration * 2},
			},
			expected: []string{
				// expect A as the C->T->A traversal timed out
				"action://a",
				// expect A as the C->A traversal timed out (this is only true if timeout was reset)
				"action://a",
			},
		},

		{
			name: "ensure parent timeout is respected at first level",
			inputs: []*testEvent{
				{name: "s"},
				{name: "t", after: timeoutDuration / 2},
				{name: "a", after: timeoutDuration * 6},
			},
			expected: []string{
				"action://s-a",
			},
		},

		{
			name: "ensure parent timeout is respected at second level",
			inputs: []*testEvent{
				{name: "s"},
				{name: "t", after: timeoutDuration / 2},
				{name: "a", after: timeoutDuration * 3},
			},
			expected: []string{
				"action://s-t-a",
			},
		},

		{
			name: "ensure parent timeout is respected outside of timeout",
			inputs: []*testEvent{
				{name: "s"},
				{name: "t", after: timeoutDuration / 2},
				{name: "a", after: timeoutDuration * 11},
			},
			expected: []string{
				"action://a",
			},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			eventMap := test.eventMap
			if eventMap == nil {
				eventMap = defaultEventMap
			}
			eventMap.Initialize()

			p, o := setupProcessor(t, eventMap)
			when := startTime
			for _, input := range test.inputs {
				when = when.Add(input.after)
				if p.previousEventTimedOut(when) {
					p.performCurrentEvent()
				}
				p.processEvent(&processorEvent{name: input.name, when: when})
			}

			p.performCurrentEvent()

			if want, got := len(test.expected), len(o); want != got {
				t.Fatalf("invalid number of events returned; want %v, got %v", want, got)
			}

			for i, expectedAction := range test.expected {
				readEvent := <-o
				if want, got := expectedAction, readEvent.Action; want != got {
					t.Errorf("unable to process event %d; want %v, got %v", i, want, got)
				}
			}
		})
	}
}
