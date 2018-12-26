package actions

import (
	"github.com/tebeka/selenium"
	"reflect"
)

type ActionBuilder struct {
	devices []interface{}
	_pointer_action *PointerActions
	_key_action *KeyActions
	driver selenium.WebDriver
}

func NewActionBuilder(driver selenium.WebDriver) *ActionBuilder{
	mouse := NewPointerInput(POINTER_MOUSE,"mouse")
	keyboard := NewKeyInput(KEY)
	input_device := []interface{}{mouse, keyboard}
	return &ActionBuilder{
		devices: input_device,
		_pointer_action: NewPointerActions(mouse),
		_key_action: NewKeyActions(keyboard),
		driver: driver,
	}
}

func (ab *ActionBuilder) add_key_input(name string) *KeyInput {
	new_input := NewKeyInput(name)
	ab._add_input(new_input)
	return new_input
}

func (ab *ActionBuilder) add_pointer_input(kind, name string) *PointerInput {
	new_input := NewPointerInput(kind, name)
	ab._add_input(new_input)
	return new_input
}

func (ab *ActionBuilder) perform() {
	enc := map[string]interface{}{"actions": []interface{}{}}
	for _, device := range ab.devices {
		reflectEncode := reflect.ValueOf(device).MethodByName("Encode")
		args := make([]reflect.Value, 0)
		encoded := reflectEncode.Call(args)
		c := encoded[0].Interface().(map[string]interface{})
		if _, ok := c["actions"]; ok {
			var list []interface{} = enc["actions"].([]interface{})
			list = append(list, c)
			enc["actions"] = list
		}
	}
	//This is done by adding a VoidCommand interface to the selenium.go and adding a VoidCommand func to the remote.go file.
	ab.driver.VoidCommand("/session/%s/actions", enc)
}

// This is done by adding a ClearActions func to the remote.go file.
func (ab *ActionBuilder) clear_actions() {
	reflectClearActions := reflect.ValueOf(ab.driver).MethodByName("ClearActions")
	args := make([]reflect.Value, 0)
	reflectClearActions.Call(args)
}

func (ab *ActionBuilder) _add_input(input interface{}) {
	if reflect.TypeOf(input).Name() == "*PointerInput" {
		input = input.(*PointerInput)
	}else if reflect.TypeOf(input).Name() == "*KeyInput" {
		input = input.(*KeyInput)
	}
	ab.devices = append(ab.devices, input)
}
