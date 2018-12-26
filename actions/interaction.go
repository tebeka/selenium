package actions

const (
	KEY = "key"
	POINTER = "pointer"
	POINTER_MOUSE = "mouse"
	POINTER_TOUCH = "touch"
	POINTER_PEN = "pen"
	PAUSE = "pause"
)
var POINTER_KINDS = [3]string{POINTER_MOUSE, POINTER_TOUCH, POINTER_PEN}
type Interaction struct {
	source interface{}
	pause string
}
func NewInteraction(sou interface{}) *Interaction {
	return &Interaction{
		source: sou,
		pause: PAUSE,
	}
}
type Pause struct {
	Interaction
	source interface{}
	duration int
}

func NewPause(sou interface{}, dur int) *Pause {
	return &Pause{
		Interaction:*NewInteraction(nil),
		source: sou,
		duration: dur,
	}
}

func (p *Pause) encode() (encoded map[string]interface{}) {
	encoded = map[string]interface{}{}
	encoded["type"] = p.Interaction.pause
	encoded["duration"] = p.duration * 1000
	return
}