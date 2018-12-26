package actions

import (
	"reflect"
)

const DEFAULT_MOVE_DURATION = 250
type PointerInput struct {
	InputDevice
	class string
	kind string
	name string
}

func NewPointerInput(kind string, name string) *PointerInput {
	findit := false
	for x := range POINTER_KINDS {
		if POINTER_KINDS[x] == kind {
			findit = true
			break
		}
	}
	if !findit {
		panic("wrong pointer kind")
	}
	return &PointerInput{
		InputDevice: *NewInputDevice(""),
		class:POINTER,
		kind:kind,
		name:name,
	}
}

func (pi *PointerInput) create_pointer_move(duration interface{}, x int, y int, origin interface{}) {
	if duration == nil {
		duration = DEFAULT_MOVE_DURATION
	}
	if reflect.ValueOf(origin).Type().String() == "*selenium.remoteWE" {
		original := map[string]string{}
		original["element-6066-11e4-a52e-4f735466cecf"] = reflect.ValueOf(origin).Elem().FieldByName("id").String()
		action := map[string]interface{}{
			"type": "pointerMove",
			"duration": duration,
			"x": x,
			"y": y,
			"origin": original,
		}
		pi.add_action(action)
	}else{
		action := map[string]interface{}{
			"type": "pointerMove",
			"duration": duration,
			"x": x,
			"y": y,
			"origin": origin,
		}
		pi.add_action(action)
	}

}

func (pi *PointerInput) Create_pointer_down(button MouseButton) {
	pi.add_action(map[string]interface{}{"type": "pointerDown", "duration": 0, "button": button})
}

func (pi *PointerInput) Create_pointer_up(button MouseButton) {
	pi.add_action(map[string]interface{}{"type": "pointerUp", "duration": 0, "button": button})
}

func (pi *PointerInput) create_pointer_cancel() {
	pi.add_action(map[string]interface{}{"type": "pointerCancel"})
}

func (pi *PointerInput) create_pause(pause_duration int) {
	pi.add_action(map[string]interface{}{"type": "pause", "duration": pause_duration * 1000})
}

func (pi *PointerInput) Encode() (encoded map[string]interface{}) {
	encoded = map[string]interface{}{}
	encoded["type"] = pi.class
	encoded["parameters"] = map[string]string{"pointerType": pi.kind}
	encoded["id"] = pi.name
	encoded["actions"] = pi.actions
	return
}
