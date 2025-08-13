package configuration

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
	"os"
	"reflect"
	"sort"
	"strconv"
	"strings"
)

// Parser 解析器可以用来解析配置文件和环境变量的定义版本，到一个统一的输出结构
type Parser struct {
	prefix  string
	mapping map[Version]VersionedParseInfo
	env     envVars
}

// NewParser 初始化 Parser 对象，给定一个环境变量前缀 prefix，以及一个特定版本的 parseInfo
func NewParser(prefix string, parseInfos []VersionedParseInfo) *Parser {
	parser := Parser{
		prefix:  prefix,
		mapping: make(map[Version]VersionedParseInfo),
	}

	// parseInfo 信息按照版本写入 mapping
	for _, parseInfo := range parseInfos {
		parser.mapping[parseInfo.Version] = parseInfo
	}

	// 记录所有的环境变量
	for _, env := range os.Environ() {
		k, v, _ := strings.Cut(env, "=")
		parser.env = append(parser.env, envVar{
			name:  k,
			value: v,
		})
	}

	// 排序，因为需要判断 storage 存不存在，才能判断 storage_aaa 存不存在
	sort.Sort(parser.env)

	return &parser
}

// Parse 读取 []byte in 内容解析到 v，如果没有配置的选项，使用环境变量进行覆盖
func (parser *Parser) Parse(in []byte, v interface{}) error {
	var versionedStruct struct {
		Version Version
	}

	if err := yaml.Unmarshal(in, &versionedStruct); err != nil {
		return err
	}

	// 判断之前创建 NewParser 时，有没有添加对应的版本信息
	parseInfo, ok := parser.mapping[versionedStruct.Version]
	if !ok { // 如果没有，则返回不支持的版本类型
		return fmt.Errorf("unsupported version: %q", versionedStruct.Version)
	}

	// 获取到当前版本的 configuration 反射类型
	parseAs := reflect.New(parseInfo.ParseAs)
	// 将配置文件转成 configuration 类型
	err := yaml.Unmarshal(in, parseAs.Interface())
	if err != nil {
		return err
	}

	// 环境变量覆写
	for _, env := range parser.env {
		pathStr := env.name
		// 判断环境变量必须要有 prefix 前缀（也就是 registry）
		if strings.HasPrefix(pathStr, strings.ToUpper(parser.prefix)+"_") {
			path := strings.Split(pathStr, "_")

			// 覆写未在配置文件中定义的配置项
			err := parser.overwriteFields(parseAs, pathStr, path[1:], env.value)
			if err != nil {
				return fmt.Errorf("parsing environment variable %s: %v", pathStr, err)
			}
		}
	}
	// 通过 conversionFunc 校验版本配置
	c, err := parseInfo.ConversionFunc(parseAs.Interface())
	if err != nil {
		return err
	}

	// reflect.ValueOf(v).Elem() 表示 v 的反设置对象的指针, Elem 如果不是指针会 panic
	// reflect.Indirect(reflect.ValueOf(c)) 表示变量c的反射值对象的指针，Indirect 比Elem()更安全，不会对非指针panic
	// 将前面Elem()获取的值设置为Indirect()返回的值
	reflect.ValueOf(v).Elem().Set(reflect.Indirect(reflect.ValueOf(c)))
	return nil
}

// overwriteFields 用于通过环境变量指定的替代值替换配置值。前提条件：绝不允许存在空路径切片传入。
// fullPath 表示环境变量完整名称
// path 表示环境变量通过 _ 切后的切片
func (parser *Parser) overwriteFields(v reflect.Value, fullPath string, path []string, payload string) error {
	// 判断 v 是否为空
	for v.Kind() == reflect.Ptr {
		if v.IsNil() {
			panic("encountered nil pointer while handling environment variable " + fullPath)
		}
		// 获取 v 的反射对象值
		v = reflect.Indirect(v)
	}

	// 根据类型进行覆盖，环境变量是 key-value，因此只能找到最下层进行覆盖，如果是 struct、map、list，需要继续往下找
	switch v.Kind() {
	case reflect.Struct:
		return parser.overwriteStructFields(v, fullPath, path, payload)
	case reflect.Map:
		return parser.overwriteMapFields(v, fullPath, path, payload)
	case reflect.Slice:
		// 如果是 slice 类型，那么首先需要指定是第几个索引
		idx, err := strconv.Atoi(path[0])
		if err != nil {
			panic("non-numeric index: " + path[0])
		}

		if idx > v.Len() { // 如果索引超出了配置的列表长度，直接报错
			panic("undefined index: " + path[0])
		}

		// 如果环境变量指定的索引不存在，那么添加一个索引，但是没有值
		if v.Len() == 0 || idx == v.Len() {
			typ := v.Type().Elem()
			elem := reflect.New(typ).Elem()
			v.Set(reflect.Append(v, elem))
		}

		// 上面添加好后，或者已经存在索引，直接更新对应索引的配置值
		return parser.overwriteFields(v.Index(idx), fullPath, path[1:], payload)
	case reflect.Interface:
		// 如果是 key=value 形式，直接更新值
		if v.NumMethod() == 0 {
			if !v.IsNil() { // 如果 v 中的对应的 key 不是空的，那么就修改其值
				return parser.overwriteFields(v.Elem(), fullPath, path, payload)
			}
			// 如果 v 中对应的 key 是空，也就是 v 是空的，那么先创建
			var template map[string]interface{}
			wrappedV := reflect.MakeMap(reflect.TypeOf(template))
			v.Set(wrappedV)
			return parser.overwriteMapFields(wrappedV, fullPath, path, payload)
		}
	}
	return nil
}

// 解析 struct 类型
func (parser *Parser) overwriteStructFields(v reflect.Value, fullPath string, path []string, payload string) error {
	byUpperCase := make(map[string]int)
	// 将 struct 类型 v 中的所有字段名，转成大写保存
	for i := 0; i < v.NumField(); i++ { // 遍历 struct 类型 v 中的所有字段
		structField := v.Type().Field(i)
		upper := strings.ToUpper(structField.Name)
		if _, exist := byUpperCase[upper]; exist { // 如果已经存在，报错重复的 key
			panic(fmt.Sprintf("field name collision in configuration object: %s", sf.Name))
		}
		byUpperCase[upper] = i
	}

	// 查找是否有环境变量中的最顶层，也就是第0个元素，例如 storage.redis.host ，最顶层就是 storage
	fieldIndex, exist := byUpperCase[path[0]]
	if !exist { // 如果环境变量不存在配置项，直接跳过
		logrus.Warnf("Ignoring unrecognized environment variable %s", fullpath)
		return nil
	}

	structField := v.Type().Field(fieldIndex)
	structFieldValue := v.Field(fieldIndex)

	// 如果环境变量层级只有一层，例如 storage，直接赋值
	if len(path) == 1 {
		newFieldVal := reflect.New(structField.Type)
		// payload 转成配置文件中对应 value 的类型
		err := yaml.Unmarshal([]byte(payload), newFieldVal.Interface())
		if err != nil {
			return err
		}
		// 赋值
		structFieldValue.Set(reflect.Indirect(newFieldVal))
		return nil
	}

	// 如果环境变量是多层级的，判断类型
	switch structField.Type.Kind() {
	case reflect.Map:
		// 如果 value 是空指针，重新初始化 map
		if structFieldValue.IsNil() {
			structFieldValue.Set(reflect.MakeMap(structField.Type))
		}
	case reflect.Ptr:
		// 如果 value 是空指针，重新初始化
		if structFieldValue.IsNil() {
			structFieldValue.Set(reflect.New(structField.Type.Elem()))
		}
	}

	err := parser.overwriteFields(structFieldValue, fullPath, path[1:], payload)
	if err != nil {
		return err
	}
	return nil
}

func (parser *Parser) overwriteMapFields(v reflect.Value, fullPath string, path []string, payload string) error {
	if v.Type().Key().Kind() != reflect.String {
		// 只支持 key 的类型为 string
		logrus.Warnf("Ignoring environment variable %s involving map with non-string keys", fullpath)
		return nil
	}

	// 如果环境变量有多层
	if len(path) > 1 {
		// 遍历 map 中所有的属性，如果能匹配到，那么直接赋值
		for _, k := range v.MapKeys() {
			if strings.ToUpper(k.String()) == path[0] {
				mapValue := v.MapIndex(k)
				// 如果当前的 value 还是空指针，那么需要后面进行初始化
				if (mapValue.Kind() == reflect.Ptr || mapValue.Kind() == reflect.Interface || mapValue.Kind() == reflect.Map) && mapValue.IsNil() {
					break
				}
				// 否则，递归进行赋值
				return parser.overwriteFields(mapValue, fullPath, path[1:], payload)
			}
		}
	}

	// 如果不存在 value ，或者 map 中不存在此 key，创建
	var mapValue reflect.Value
	if v.Type().Elem().Kind() == reflect.Map {
		mapValue = reflect.MakeMap(v.Type().Elem())
	} else {
		mapValue = reflect.New(v.Type().Elem())
	}

	if len(path) > 1 {
		err := parser.overwriteFields(mapValue, fullPath, path[1:], payload)
		if err != nil {
			return err
		}
	} else { // 环境变量只有一层
		err := yaml.Unmarshal([]byte(payload), mapValue.Interface())
		if err != nil {
			return err
		}
	}

	// 赋值
	v.SetMapIndex(reflect.ValueOf(strings.ToLower(path[0])), reflect.Indirect(mapValue))

	return nil
}
