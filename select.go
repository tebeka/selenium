package selenium

import (
	"fmt"
	"strings"
)

// SelectElement WebElement that is specific to the Select Dropdown
type SelectElement struct {
	element WebElement
	isMulti bool
}

// Select Creates a SelectElement
// @param el The initial WebElement
func Select(el WebElement) (se SelectElement, err error) {
	se = SelectElement{}

	tagName, err := el.TagName()
	if err != nil || strings.ToLower(tagName) != "select" {
		err = fmt.Errorf(`element should have been "select" but was "%s"`, tagName)
		return
	}

	se.element = el
	mult, err2 := el.GetAttribute("multiple")
	se.isMulti = (err2 != nil && strings.ToLower(mult) != "false")

	return
}

// GetElement Gets the raw WebElement
func (s SelectElement) GetElement() WebElement {
	return s.element
}

// IsMultiple Whether this select element support selecting multiple options at the same time? This
//            is done by checking the value of the "multiple" attribute.
func (s SelectElement) IsMultiple() bool {
	return s.isMulti
}

// GetOptions Returns all of the options of that Select
func (s SelectElement) GetOptions() ([]WebElement, error) {
	return s.element.FindElements(ByTagName, "option")
}

// GetAllSelectedOptions Returns all of the options of that Select that are selected
func (s SelectElement) GetAllSelectedOptions() ([]WebElement, error) {
	// return getOptions().stream().filter(WebElement::isSelected).collect(Collectors.toList());

	var opts []WebElement
	return opts, nil
}

// GetFirstSelectedOption Returns the first selected option of the Select Element
func (s SelectElement) GetFirstSelectedOption() (opt WebElement, err error) {
	opts, err := s.GetAllSelectedOptions()
	if err != nil {
		return
	}
	opt = opts[0]
	return
}

// SelectByVisibleText Select all options that display text matching the argument. That is,
//                      when given "Bar" this would select an option like:
//
// <option value="foo">Bar</option>
//
// @param text The visible text to match against
//
func (s SelectElement) SelectByVisibleText(text string) error {
	// try to find the option via XPATH ...
	options, err := s.element.FindElements(ByXPATH, `.//option[normalize-space(.) = "`+escapeQuotes(text)+`"]`)
	if err != nil {
		return err
	}

	for _, option := range options {
		s.setSelected(option, true)
		if !s.isMulti {
			return nil
		}
	}

	matched := len(options) > 0
	if !matched && strings.Contains(text, " ") {
		subStringWithoutSpace := getLongestSubstringWithoutSpace(text)
		var candidates []WebElement
		if subStringWithoutSpace == "" {
			// hmm, text is either empty or contains only spaces - get all options ...
			candidates, err = s.GetOptions()
		} else {
			// get candidates via XPATH ...
			candidates, err = s.element.FindElements(ByXPATH, `.//option[contains(., "`+escapeQuotes(subStringWithoutSpace)+`")]`)
		}

		if err != nil {
			return err
		}

		trimmed := strings.TrimSpace(text)

		for _, option := range candidates {
			o, err := option.Text()
			if err != nil {
				return err
			}
			if trimmed == strings.TrimSpace(o) {
				s.setSelected(option, true)
				if !s.isMulti {
					return nil
				}
				matched = true
			}
		}
	}
	if !matched {
		return fmt.Errorf("cannot locate option with text: %s", text)
	}
	return nil
}

// SelectByIndex Select the option at the given index. This is done by examining the "index" attribute of an
//               element, and not merely by counting.
//
// @param idx The option at this index will be selected
func (s SelectElement) SelectByIndex(idx int) error {
	return s.setSelectedByIndex(idx, true)
}

// SelectByValue Select all options that have a value matching the argument. That is, when given "foo" this
//               would select an option like:
//
//               <option value="foo">Bar</option>
//
// @param value The value to match against

func (s SelectElement) SelectByValue(value string) error {
	opts, err := s.findOptionsByValue(value)
	if err != nil {
		return err
	}
	for _, option := range opts {
		s.setSelected(option, true)
		if !s.isMulti {
			return nil
		}
	}
	return nil
}

// DeselectAll Clear all selected entries. This is only valid when the SELECT supports multiple selections.
func (s SelectElement) DeselectAll() error {
	if !s.isMulti {
		return fmt.Errorf("you may only deselect all options of a multi-select")
	}

	opts, err := s.GetOptions()
	if err != nil {
		return err
	}
	for _, o := range opts {
		err = s.setSelected(o, false)
		if err != nil {
			return err
		}
	}
	return nil
}

// DeselectByValue Deselect all options that have a value matching the argument. That is, when given "foo" this
//                 would deselect an option like:
//
// <option value="foo">Bar</option>
//
// @param value The value to match against
func (s SelectElement) DeselectByValue(value string) error {
	if !s.isMulti {
		return fmt.Errorf("you may only deselect all options of a multi-select")
	}

	opts, err := s.findOptionsByValue(value)
	if err != nil {
		return err
	}
	for _, o := range opts {
		err = s.setSelected(o, false)
		if err != nil {
			return err
		}
	}
	return nil
}

// DeselectByIndex Deselect the option at the given index. This is done by examining the "index" attribute of an
//                 element, and not merely by counting.
//
// @param index The option at this index will be deselected
func (s SelectElement) DeselectByIndex(index int) error {
	if !s.isMulti {
		return fmt.Errorf("you may only deselect all options of a multi-select")
	}

	return s.setSelectedByIndex(index, false)
}

// DeselectByVisibleText Deselect all options that display text matching the argument. That is,
//                       when given "Bar" this would deselect an option like:
//
// <option value="foo">Bar</option>
//
// @param text The visible text to match against
func (s SelectElement) DeselectByVisibleText(text string) error {
	if !s.isMulti {
		return fmt.Errorf("you may only deselect all options of a multi-select")
	}

	options, err := s.element.FindElements(ByXPATH, `.//option[normalize-space(.) = "`+escapeQuotes(text)+`"]`)
	if err != nil {
		return err
	}
	if len(options) == 0 {
		return fmt.Errorf("Cannot locate option with text: " + text)
	}

	for _, option := range options {
		err = s.setSelected(option, false)
		if err != nil {
			return err
		}
	}
	return nil
}

func escapeQuotes(str string) string {
	str1 := strings.Replace(str, `"`, `\"`, -1)
	return str1
}

func getLongestSubstringWithoutSpace(s string) string {
	result := ""
	st := strings.Split(s, " ")
	for _, t := range st {
		if len(t) > len(result) {
			result = t
		}
	}
	return result
}

func (s SelectElement) findOptionsByValue(value string) (opts []WebElement, err error) {
	opts, err = s.element.FindElements(ByXPATH, `.//option[@value = "`+escapeQuotes(value)+`"]`)
	if err != nil {
		return
	}
	if len(opts) == 0 {
		err = fmt.Errorf("Cannot locate option with value: " + value)
	}

	return
}

func (s SelectElement) setSelectedByIndex(index int, selected bool) error {
	idx := fmt.Sprintf("%d", index)
	opts, err := s.element.FindElements(ByXPATH, `.//option[@index = "`+idx+`"]`)
	if err != nil {
		return err
	}
	if len(opts) == 0 {
		err = fmt.Errorf("Cannot locate option with index: " + idx)
		return err
	}

	err = s.setSelected(opts[index], selected)

	return err
}

func (s SelectElement) setSelected(option WebElement, selected bool) (err error) {
	sel, err := option.IsSelected()
	if sel != selected && err == nil {
		err = option.Click()
	}
	return err
}
