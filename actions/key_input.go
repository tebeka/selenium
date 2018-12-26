package actions

type KeyInput struct {
	InputDevice
	class string
	name string
}

func NewKeyInput(name string) *KeyInput {
	return &KeyInput{
		InputDevice: *NewInputDevice(""),
		class:KEY,
		name:name,
	}
}

func (ki *KeyInput) create_key_down(key string) {
	ki.add_action(map[string]interface{}{"type": "keyDown", "value": key})
}

func (ki *KeyInput) create_key_up(key string) {
	ki.add_action(map[string]interface{}{"type": "keyUp", "value": key})
}

func (ki *KeyInput) create_pause(pause_duration int) {
	ki.add_action(map[string]interface{}{"type": "pause", "duration": pause_duration * 1000})
}

func (ki *KeyInput) Encode() (encoded map[string]interface{}) {
	encoded = map[string]interface{}{}
	encoded["type"] = ki.class
	encoded["id"] = ki.name
	encoded["actions"] = ki.actions
	return
}
