// The user action events.
package event

type Event int

const (
	NOP Event = iota

	// the click operation
	CLICK
	DOUBLE_CLICK

	// slide the display to the next page
	SLIDE_UP
	SLIDE_DOWN
	SLIDE_LEFT
	SLIDE_RIGHT
)

// Show the human-readable string of the event.
func (e Event) String() string {
	switch e {
	case CLICK:
		return "click"
	case DOUBLE_CLICK:
		return "double click"
	case SLIDE_UP:
		return "slide up"
	case SLIDE_DOWN:
		return "slide down"
	case SLIDE_LEFT:
		return "slide left"
	case SLIDE_RIGHT:
		return "slide right"
	default:
		return "UNKNOWN"
	}
}
