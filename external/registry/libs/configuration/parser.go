package configuration

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
	"reflect"
	"sort"
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

	reflect.ValueOf(v).Elem().Set(reflect.Indirect(reflect.ValueOf(c)))
	return nil
}

// overwriteFields 用于通过环境变量指定的替代值替换配置值。前提条件：绝不允许存在空路径切片传入。
func (parser *Parser) overwriteFields(v reflect.Value, fullPath string, path []string, payload string) error {
	return nil
}
