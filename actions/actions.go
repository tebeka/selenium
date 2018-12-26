package actions

import (
	"github.com/tebeka/selenium"
	"reflect"
	"time"
)

type Actions struct {
	_driver selenium.WebDriver
	_actions []func()
	w3c_actions *ActionBuilder
	w3c bool
}

func (ac *Actions) NewActionChains() *Actions{
	w3cCompatible := reflect.ValueOf(ac._driver).Elem().FieldByName("w3cCompatible")
	ac.w3c = w3cCompatible.Bool()
	if ac.w3c {
		ac.w3c_actions = NewActionBuilder(ac._driver)
	}
	return ac
}

func (ac *Actions) Perform() {
	if ac.w3c {
		ac.w3c_actions.perform()
	}else {
		for _, v := range ac._actions {
			v()
		}
	}
}

func (ac *Actions) Reset_actions() {
	if ac.w3c {
		ac.w3c_actions.clear_actions()
	}
	ac._actions = []func(){}
}

func (ac *Actions) Move_to_element(to_element selenium.WebElement) *Actions{
	if ac.w3c {
		ac.w3c_actions._pointer_action.move_to(to_element, 0, 0)
		ac.w3c_actions._key_action.pause(0)
	}else {
		ac._actions = append(ac._actions, func(){
			to_element.MoveTo(0, 0)
		})
	}
	return ac
}

func (ac *Actions) Move_by_offset(xoffset, yoffset int) *Actions{
	if ac.w3c {
		ac.w3c_actions._pointer_action.move_by(xoffset, yoffset)
		ac.w3c_actions._key_action.pause(0)
	}else {
		// This is done by adding a MoveBy func to the remote.go file and adding a MoveBy interface to the selenium.go 
		ac._actions = append(ac._actions, func(){
			ac._driver.MoveBy(xoffset, yoffset)
		})
	}
	return ac
}

func (ac *Actions) Move_to_element_with_offset(to_element selenium.WebElement, xoffset, yoffset int) *Actions{
	if ac.w3c {
		ac.w3c_actions._pointer_action.move_to(to_element, xoffset, yoffset)
		ac.w3c_actions._key_action.pause(0)
	}else {
		ac._actions = append(ac._actions, func(){
			to_element.MoveTo(xoffset, yoffset)
		})
	}
	return ac
}

func (ac *Actions) Click(on_element interface{}) *Actions {
	if _, ok := on_element.(selenium.WebElement); ok {
		ac.Move_to_element(on_element.(selenium.WebElement))
	}
	if ac.w3c {
		ac.w3c_actions._pointer_action.click(nil)
		ac.w3c_actions._key_action.pause(0)
		ac.w3c_actions._key_action.pause(0)
	}else {
		ac._actions = append(ac._actions, func(){
			ac._driver.Click(selenium.LeftButton)
		})
	}
	return ac
}

func (ac *Actions) Click_and_hold(on_element interface{}) *Actions {
	if _, ok := on_element.(selenium.WebElement); ok {
		ac.Move_to_element(on_element.(selenium.WebElement))
	}
	if ac.w3c {
		ac.w3c_actions._pointer_action.click_and_hold(nil)
		ac.w3c_actions._key_action.pause(0)
	}else {
		ac._actions = append(ac._actions, func(){
			ac._driver.ButtonDown()
		})
	}
	return ac
}

func (ac *Actions) Context_click(on_element interface{}) *Actions {
	if _, ok := on_element.(selenium.WebElement); ok {
		ac.Move_to_element(on_element.(selenium.WebElement))
	}
	if ac.w3c {
		ac.w3c_actions._pointer_action.context_click(nil)
		ac.w3c_actions._key_action.pause(0)
		ac.w3c_actions._key_action.pause(0)
	}else {
		ac._actions = append(ac._actions, func(){
			ac._driver.Click(selenium.RightButton)
		})
	}
	return ac
}

func (ac *Actions) Double_click(on_element interface{}) *Actions {
	if _, ok := on_element.(selenium.WebElement); ok {
		ac.Move_to_element(on_element.(selenium.WebElement))
	}
	if ac.w3c {
		ac.w3c_actions._pointer_action.double_click(nil)
		ac.w3c_actions._key_action.pause(0)
		ac.w3c_actions._key_action.pause(0)
		ac.w3c_actions._key_action.pause(0)
		ac.w3c_actions._key_action.pause(0)
	}else {
		ac._actions = append(ac._actions, func(){
			ac._driver.DoubleClick()
		})
	}
	return ac
}

func (ac *Actions) Drag_and_drop(source, target selenium.WebElement) *Actions {
	ac.Click_and_hold(source)
	ac.Release(target)
	return ac
}

func (ac *Actions) Drag_and_drop_by_offset(source selenium.WebElement, xoffset, yoffset int) *Actions {
	ac.Click_and_hold(source)
	ac.Move_by_offset(xoffset, yoffset)
	ac.Release(nil)
	return ac
}

func (ac *Actions) Key_down(text string, element selenium.WebElement) *Actions {
	if _, ok := element.(selenium.WebElement); ok {
		ac.Click(element)
	}
	if ac.w3c {
		ac.w3c_actions._key_action.key_down(text)
		ac.w3c_actions._pointer_action.pause(0)
	}else {
		if _, ok := element.(selenium.WebElement); ok {
			element.SendKeys(text)
		}else {
			ae, _ := ac._driver.ActiveElement()
			ae.SendKeys(text)
		}
	}
	return ac
}

func (ac *Actions) Key_up(text string, element selenium.WebElement) *Actions {
	if _, ok := element.(selenium.WebElement); ok {
		ac.Click(element)
	}
	if ac.w3c {
		ac.w3c_actions._key_action.key_up(text)
		ac.w3c_actions._pointer_action.pause(0)
	}else {
		if _, ok := element.(selenium.WebElement); ok {
			element.SendKeys(text)
		}else {
			ae, _ := ac._driver.ActiveElement()
			ae.SendKeys(text)
		}
	}
	return ac
}

func (ac *Actions) Send_keys(text string) *Actions {
	if ac.w3c{
		if len(text) > 1 {
			for _, c := range text {
				ac.Key_down(string(c), nil)
				ac.Key_up(string(c), nil)
			}
		}else {
			ac.Key_down(text, nil)
			ac.Key_up(text, nil)
		}
	}else {
		ae, _ := ac._driver.ActiveElement()
		ae.SendKeys(text)
	}
	return ac
}

func (ac *Actions) Send_keys_to_element(element selenium.WebElement, text string) *Actions {
	ac.Click(element)
	ac.Send_keys(text)
	return ac
}

func (ac *Actions) Pause(seconds int) *Actions {
	if ac.w3c{
		ac.w3c_actions._pointer_action.pause(seconds)
		ac.w3c_actions._key_action.pause(seconds)
	}else {
		ac._actions = append(ac._actions, func(){
			time.Sleep(time.Duration(seconds) * time.Second)
		})
	}
	return ac
}

func (ac *Actions) Release(on_element interface{}) *Actions {
	if _, ok := on_element.(selenium.WebElement); ok {
		ac.Move_to_element(on_element.(selenium.WebElement))
	}
	if ac.w3c {
		ac.w3c_actions._pointer_action.release()
		ac.w3c_actions._key_action.pause(0)

	}else {
		ac._actions = append(ac._actions, func(){
			ac._driver.ButtonUp()
		})
	}
	return ac
}

func ActionChains(driver selenium.WebDriver) *Actions{
	ac := Actions{
		_driver: driver,
		_actions: []func(){},
	}
	return ac.NewActionChains()
}
