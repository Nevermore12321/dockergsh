package engine

type Hack map[string]interface{}

func (eng *Engine) Hack_SetGlobalVar(key string, value interface{}) {
	if eng.hack == nil {
		eng.hack = make(Hack)
	}
	eng.hack[key] = value
}

func (eng *Engine) Hack_GetGlobalVar(key string) interface{} {
	if eng == nil {
		return nil
	}
	val, exists := eng.hack[key]
	if !exists {
		return nil
	}
	return val
}
