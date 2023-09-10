package engine

import "strings"

// Env 就是每个 job 所需要的 环境变量设置
type Env []string

// 统一设置 key=value 环境变量，都是 string 类型，格式为 key=value 列表
func (env *Env) Set(key, value string) {
	*env = append(*env, key+"="+value)
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
