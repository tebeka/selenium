package actions

import (
	"fmt"
	"github.com/satori/go.uuid"
)

type InputDevice struct {
	name string
	actions []map[string]interface{}
}

func NewInputDevice(deviceName string) *InputDevice {
	if deviceName == "" {
		deviceName = fmt.Sprintf("%s", uuid.Must(uuid.NewV4()))
	}
	return &InputDevice{
		name: deviceName,
		actions: []map[string]interface{}{},
	}
}

func (device *InputDevice) add_action(action map[string]interface{}) {
	device.actions = append(device.actions, action)
	return
}

func (device *InputDevice) clear_actions() {
	device.actions = device.actions[0:0]
}