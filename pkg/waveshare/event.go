// Convert the touch event to the PwnPi event.
package waveshare

import (
	"context"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/cmj0121/pwnpi/pkg/event"
)

const (
	// Consider the click events
	DURATION_DOUBLE_CLICK = 300 * time.Millisecond
	DURATION_TWO_EVENT    = 1200 * time.Millisecond

	// The threshold of the touch event to slide up/down or left/right
	THRESHOLD_SLIDE_UP_DOWN    = 10.0
	THRESHOLD_SLIDE_LEFT_RIGHT = 0.5
)

// The event reader to handle the touch event.
func (i *I2C) Event(ctx context.Context) <-chan event.Event {
	ch := make(chan event.Event)

	go func() {
		defer close(ch)

		var prev *MultiTouch
		toucher := i.Touches(ctx)

		for {
			select {
			case <-ctx.Done():
				return
			case touch := <-toucher:
				switch e := i.convertToEvent(prev, &touch); e {
				case event.NOP:
					continue
				default:
					ch <- e
					prev = &touch
				}
			}
		}
	}()

	return ch
}

// Convert the two click events to the PwnPi event.
func (i *I2C) convertToEvent(prev, curr *MultiTouch) event.Event {
	log.Trace().Interface("prev", prev).Interface("curr", curr).Msg("convert multi-touch to event")

	switch {
	case prev == nil && curr == nil:
		// both are nil, no touch event
		return event.NOP
	case prev == nil:
		// the first touch event
		return event.CLICK
	case prev.Timestamp.Add(DURATION_DOUBLE_CLICK).Before(curr.Timestamp):
		// consider as the double click
		return event.DOUBLE_CLICK
	}

	// handle the complex touch event
	x, y, r := prev.MoveTo(curr)
	switch {
	case r >= THRESHOLD_SLIDE_UP_DOWN:
		switch {
		case y < 0:
			return event.SLIDE_UP
		case y > 0:
			return event.SLIDE_DOWN
		}
	case r <= THRESHOLD_SLIDE_LEFT_RIGHT:
		switch {
		case x < 0:
			return event.SLIDE_RIGHT
		case x > 0:
			return event.SLIDE_LEFT
		}
	}

	return event.NOP
}
