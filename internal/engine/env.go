package engine

import (
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// Env 就是每个 job 所需要的 环境变量设置
// 格式就是 KEY=VALUE
type Env []string

// Map 将环境变量 key=value 形式转成 map 返回
func (env Env) Map() map[string]string {
	m := make(map[string]string)
	for _, e := range env {
		pair := strings.SplitN(e, "=", 2)
		if len(pair) != 2 {
			continue
		}
		m[pair[0]] = pair[1]
	}
	return m
}

// Set 统一设置 key=value 环境变量，都是 string 类型，格式为 key=value 列表
func (env *Env) Set(key, value string) {
	*env = append(*env, key+"="+value)
}

func (env *Env) SetInt64(key string, value int64) {
	env.Set(key, strconv.FormatInt(value, 10))
}

func (env *Env) Get(key string) (value string) {
	for _, kv := range *env {
		if strings.Index(kv, "=") == -1 {
			continue
		}
		parts := strings.SplitN(kv, "=", 2)
		if parts[0] != key {
			continue
		}
		if len(parts) < 2 {
			value = ""
		} else {
			value = parts[1]
		}
	}
	return
}

func (env *Env) Exists(key string) bool {
	_, exists := env.Map()[key]
	return exists
}

func (env *Env) Len() int {
	return len(env.Map())
}

func (env *Env) SetBool(key string, value bool) {
	if value {
		env.Set(key, "1")
	} else {
		env.Set(key, "0")
	}
}

func (env *Env) GetBool(key string) (value bool) {
	s := strings.ToLower(strings.Trim(env.Get(key), "\t"))
	if s == "" || s == "0" || s == "no" || s == "false" || s == "none" {
		value = false
	} else {
		value = true
	}
	return
}

// Encode 将 env 中的 value 转成对应的类型
func (env *Env) Encode(dst io.Writer) error {
	m := make(map[string]interface{})
	for k, v := range env.Map() {
		var val interface{}
		if err := json.Unmarshal([]byte(v), &val); err == nil {
			m[k] = changeFloats(val)
		} else {
			m[k] = val
		}
	}

	if err := json.NewEncoder(dst).Encode(&m); err != nil {
		return err
	}
	return nil
}

// Decode 将 src 中的数据转成 env 的 map
func (env *Env) Decode(src io.Reader) error {
	m := make(map[string]interface{})
	if err := json.NewDecoder(src).Decode(&m); err != nil {
		return err
	}
	for k, v := range m {
		env.SetAuto(k, v)
	}
	return nil
}

func (env *Env) SetAuto(key string, value interface{}) {
	if fval, ok := value.(float64); ok {
		env.SetInt64(key, int64(fval))
	} else if sval, ok := value.(string); ok {
		env.Set(key, sval)
	} else if val, err := json.Marshal(value); err == nil {
		env.Set(key, string(val))
	} else {
		env.Set(key, fmt.Sprintf("%v", value))
	}
}

func changeFloats(v interface{}) interface{} {
	switch v := v.(type) {
	case float64:
		return int(v)
	case map[string]interface{}:
		for key, val := range v {
			v[key] = changeFloats(val)
		}
	case []interface{}:
		for i, val := range v {
			v[i] = changeFloats(val)
		}
	}
	return v
}

// Decoder 自定义的 json decode
type Decoder struct {
	*json.Decoder
}

func (decoder *Decoder) Decode() (*Env, error) {
	m := make(map[string]interface{})
	if err := decoder.Decoder.Decode(&m); err != nil {
		return nil, err
	}

	env := &Env{}
	for k, v := range m {
		env.SetAuto(k, v)
	}

	return env, nil
}
