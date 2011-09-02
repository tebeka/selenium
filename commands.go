package selenium

import (
	"bytes"
	"os"
	"template"
)

type Params struct {
	Id, Name, Other, PropertyName, SessionId string
}

type Command struct {
	Method string
	urlTemplate *template.Template
}

func (cmd *Command) URL(params *Params) (string, os.Error) {
	var buf bytes.Buffer
	if err := cmd.urlTemplate.Execute(&buf, params); err != nil {
		return "", err
	}

	return string(buf.String()), nil
}

var CMD_STATUS = &Command{"GET", template.MustParse("/status", nil)}
var CMD_NEW_SESSION = &Command{"POST", template.MustParse("/session", nil)}
var CMD_QUIT = &Command{"DELETE", template.MustParse("/session/{SessionId}", nil)}
var CMD_GET_CURRENT_WINDOW_HANDLE = &Command{"GET", template.MustParse("/session/{SessionId}/window_handle", nil)}
var CMD_GET_WINDOW_HANDLES = &Command{"GET", template.MustParse("/session/{SessionId}/window_handles", nil)}
var CMD_GET = &Command{"POST", template.MustParse("/session/{SessionId}/url", nil)}
var CMD_GO_FORWARD = &Command{"POST", template.MustParse("/session/{SessionId}/forward", nil)}
var CMD_GO_BACK = &Command{"POST", template.MustParse("/session/{SessionId}/back", nil)}
var CMD_REFRESH = &Command{"POST", template.MustParse("/session/{SessionId}/refresh", nil)}
var CMD_EXECUTE_SCRIPT = &Command{"POST", template.MustParse("/session/{SessionId}/execute", nil)}
var CMD_GET_CURRENT_URL = &Command{"GET", template.MustParse("/session/{SessionId}/url", nil)}
var CMD_GET_TITLE = &Command{"GET", template.MustParse("/session/{SessionId}/title", nil)}
var CMD_GET_PAGE_SOURCE = &Command{"GET", template.MustParse("/session/{SessionId}/source", nil)}
var CMD_SCREENSHOT = &Command{"GET", template.MustParse("/session/{SessionId}/screenshot", nil)}
var CMD_SET_BROWSER_VISIBLE = &Command{"POST", template.MustParse("/session/{SessionId}/visible", nil)}
var CMD_IS_BROWSER_VISIBLE = &Command{"GET", template.MustParse("/session/{SessionId}/visible", nil)}
var CMD_FIND_ELEMENT = &Command{"POST", template.MustParse("/session/{SessionId}/element", nil)}
var CMD_FIND_ELEMENTS = &Command{"POST", template.MustParse("/session/{SessionId}/elements", nil)}
var CMD_GET_ACTIVE_ELEMENT = &Command{"POST", template.MustParse("/session/{SessionId}/element/active", nil)}
var CMD_FIND_CHILD_ELEMENT = &Command{"POST", template.MustParse("/session/{SessionId}/element/{Id}/element", nil)}
var CMD_FIND_CHILD_ELEMENTS = &Command{"POST", template.MustParse("/session/{SessionId}/element/{Id}/elements", nil)}
var CMD_CLICK_ELEMENT = &Command{"POST", template.MustParse("/session/{SessionId}/element/{Id}/click", nil)}
var CMD_CLEAR_ELEMENT = &Command{"POST", template.MustParse("/session/{SessionId}/element/{Id}/clear", nil)}
var CMD_SUBMIT_ELEMENT = &Command{"POST", template.MustParse("/session/{SessionId}/element/{Id}/submit", nil)}
var CMD_GET_ELEMENT_TEXT = &Command{"GET", template.MustParse("/session/{SessionId}/element/{Id}/text", nil)}
var CMD_SEND_KEYS_TO_ELEMENT = &Command{"POST", template.MustParse("/session/{SessionId}/element/{Id}/value", nil)}
var CMD_SEND_MODIFIER_KEY_TO_ACTIVE_ELEMENT = &Command{"POST", template.MustParse("/session/{SessionId}/modifier", nil)}
var CMD_GET_ELEMENT_VALUE = &Command{"GET", template.MustParse("/session/{SessionId}/element/{Id}/value", nil)}
var CMD_GET_ELEMENT_TAG_NAME = &Command{"GET", template.MustParse("/session/{SessionId}/element/{Id}/name", nil)}
var CMD_IS_ELEMENT_SELECTED = &Command{"GET", template.MustParse("/session/{SessionId}/element/{Id}/selected", nil)}
var CMD_SET_ELEMENT_SELECTED = &Command{"POST", template.MustParse("/session/{SessionId}/element/{Id}/selected", nil)}
var CMD_TOGGLE_ELEMENT = &Command{"POST", template.MustParse("/session/{SessionId}/element/{Id}/toggle", nil)}
var CMD_IS_ELEMENT_ENABLED = &Command{"GET", template.MustParse("/session/{SessionId}/element/{Id}/enabled", nil)}
var CMD_IS_ELEMENT_DISPLAYED = &Command{"GET", template.MustParse("/session/{SessionId}/element/{Id}/displayed", nil)}
var CMD_HOVER_OVER_ELEMENT = &Command{"POST", template.MustParse("/session/{SessionId}/element/{Id}/hover", nil)}
var CMD_GET_ELEMENT_LOCATION = &Command{"GET", template.MustParse("/session/{SessionId}/element/{Id}/location", nil)}
var CMD_GET_ELEMENT_LOCATION_ONCE_SCROLLED_INTO_VIEW = &Command{"GET", template.MustParse("/session/{SessionId}/element/{Id}/location_in_view", nil)}
var CMD_GET_ELEMENT_SIZE = &Command{"GET", template.MustParse("/session/{SessionId}/element/{Id}/size", nil)}
var CMD_GET_ELEMENT_ATTRIBUTE = &Command{"GET", template.MustParse("/session/{SessionId}/element/{Id}/attribute/{Name}", nil)}
var CMD_ELEMENT_EQUALS = &Command{"GET", template.MustParse("/session/{SessionId}/element/{Id}/equals/{Other}", nil)}
var CMD_GET_ALL_COOKIES = &Command{"GET", template.MustParse("/session/{SessionId}/cookie", nil)}
var CMD_ADD_COOKIE = &Command{"POST", template.MustParse("/session/{SessionId}/cookie", nil)}
var CMD_DELETE_ALL_COOKIES = &Command{"DELETE", template.MustParse("/session/{SessionId}/cookie", nil)}
var CMD_DELETE_COOKIE = &Command{"DELETE", template.MustParse("/session/{SessionId}/cookie/{Name}", nil)}
var CMD_SWITCH_TO_FRAME = &Command{"POST", template.MustParse("/session/{SessionId}/frame", nil)}
var CMD_SWITCH_TO_WINDOW = &Command{"POST", template.MustParse("/session/{SessionId}/window", nil)}
var CMD_CLOSE = &Command{"DELETE", template.MustParse("/session/{SessionId}/window", nil)}
var CMD_DRAG_ELEMENT = &Command{"POST", template.MustParse("/session/{SessionId}/element/{Id}/drag", nil)}
var CMD_GET_SPEED = &Command{"GET", template.MustParse("/session/{SessionId}/speed", nil)}
var CMD_SET_SPEED = &Command{"POST", template.MustParse("/session/{SessionId}/speed", nil)}
var CMD_GET_ELEMENT_VALUE_OF_CSS_PROPERTY = &Command{"GET", template.MustParse("/session/{SessionId}/element/{Id}/css/{PropertyName}", nil)}
var CMD_IMPLICIT_WAIT = &Command{"POST", template.MustParse("/session/{SessionId}/timeouts/implicit_wait", nil)}
var CMD_EXECUTE_ASYNC_SCRIPT = &Command{"POST", template.MustParse("/session/{SessionId}/execute_async", nil)}
var CMD_SET_SCRIPT_TIMEOUT = &Command{"POST", template.MustParse("/session/{SessionId}/timeouts/async_script", nil)}
var CMD_DISMISS_ALERT = &Command{"POST", template.MustParse("/session/{SessionId}/dismiss_alert", nil)}
var CMD_ACCEPT_ALERT = &Command{"POST", template.MustParse("/session/{SessionId}/accept_alert", nil)}
var CMD_SET_ALERT_VALUE = &Command{"POST", template.MustParse("/session/{SessionId}/alert_text", nil)}
var CMD_GET_ALERT_TEXT = &Command{"GET", template.MustParse("/session/{SessionId}/alert_text", nil)}
var CMD_CLICK = &Command{"POST", template.MustParse("/session/{SessionId}/click", nil)}
var CMD_DOUBLE_CLICK = &Command{"POST", template.MustParse("/session/{SessionId}/doubleclick", nil)}
var CMD_MOUSE_DOWN = &Command{"POST", template.MustParse("/session/{SessionId}/buttondown", nil)}
var CMD_MOUSE_UP = &Command{"POST", template.MustParse("/session/{SessionId}/buttonup", nil)}
var CMD_MOVE_TO = &Command{"POST", template.MustParse("/session/{SessionId}/moveto", nil)}

