package actions

import (
	"github.com/tebeka/selenium"
	"reflect"
)

type PointerActions struct {
	Interaction
	source *PointerInput
}

func NewPointerActions(source interface{}) *PointerActions {
	if source == nil {
		source = *NewPointerInput(POINTER_MOUSE, "mouse")
	}
	return &PointerActions{
		Interaction: *NewInteraction(source),
		source: source.(*PointerInput),
	}
}

func (pa *PointerActions) pointer_down(button MouseButton) {
	pa._button_action("Create_pointer_down", button)
}

func (pa *PointerActions) pointer_up(button MouseButton) {
	pa._button_action("Create_pointer_up", button)
}

func (pa *PointerActions) move_to(element selenium.WebElement, x, y int) *PointerActions {
	var left, top int
	if x == 0 && y == 0{
		left = 0
		top = 0
	}else{
		//This is done by adding a Rect interface to the selenium.go and adding a Rect func to the remote.go file.
		el_rect, err:= element.Rect()
		if err != nil {
			panic("wrong element rect")
		}
		left_offset := int(el_rect.Width / 2)
		top_offset := int(el_rect.Height / 2)
		left = -left_offset + x
		top = -top_offset + y
	}
	pa.source.create_pointer_move(nil, left, top, element)
	return pa
}

func (pa *PointerActions) move_by(x, y int) *PointerActions {
	pa.source.create_pointer_move(nil, x, y, POINTER)
	return pa
}

func (pa *PointerActions) move_to_location(x, y int) *PointerActions {
	pa.source.create_pointer_move(nil, x, y, "viewport")
	return pa
}

func (pa *PointerActions) click(element interface{}) *PointerActions {
	if _, ok := element.(selenium.WebElement); ok {
		pa.move_to(element.(selenium.WebElement), 0, 0)
	}
	pa.pointer_down(MouseButtonLeft)
	pa.pointer_up(MouseButtonLeft)
	return pa
}

func (pa *PointerActions) context_click(element interface{}) *PointerActions {
	if _, ok := element.(selenium.WebElement); ok {
		pa.move_to(element.(selenium.WebElement), 0, 0)
	}
	pa.pointer_down(MouseButtonRight)
	pa.pointer_up(MouseButtonRight)
	return pa
}

func (pa *PointerActions) click_and_hold(element interface{}) *PointerActions {
	if _, ok := element.(selenium.WebElement); ok {
		pa.move_to(element.(selenium.WebElement), 0, 0)
	}
	pa.pointer_down(MouseButtonLeft)
	return pa
}

func (pa *PointerActions) release() *PointerActions {
	pa.pointer_up(MouseButtonLeft)
	return pa
}

func (pa *PointerActions) double_click(element interface{}) *PointerActions {
	if _, ok := element.(selenium.WebElement); ok {
		pa.move_to(element.(selenium.WebElement), 0, 0)
	}
	pa.click(nil)
	pa.click(nil)
	return pa
}

func (pa *PointerActions) pause(duration int) *PointerActions {
	pa.source.create_pause(duration)
	return pa
}

func (pa *PointerActions) _button_action(action string, button MouseButton) *PointerActions {
	v := reflect.ValueOf(pa.source)
	ac := v.MethodByName(action)
	args := []reflect.Value{reflect.ValueOf(button)}
	ac.Call(args)
	return pa
 }
