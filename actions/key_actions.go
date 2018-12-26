package actions

type KeyActions struct {
	Interaction
	source *KeyInput
}

func NewKeyActions(source interface{}) *KeyActions {
	if source == nil {
		source = *NewKeyInput(KEY)
	}
	return &KeyActions{
		Interaction: *NewInteraction(source),
		source: source.(*KeyInput),
	}
}

func (ka *KeyActions) key_down(key string) {
	ka.source.create_key_down(key)
}

func (ka *KeyActions) key_up(key string) {
	ka.source.create_key_up(key)
}

func (ka *KeyActions) pause(duration int) *KeyActions {
	ka.source.create_pause(duration)
	return ka
}

func (ka *KeyActions) send_keys(text string) *KeyActions {
	if len(text) > 1 {
		for _, c := range text {
			ka.key_down(string(c))
			ka.key_up(string(c))
		}
	}else {
		ka.key_down(text)
		ka.key_up(text)
	}
	return ka
}
