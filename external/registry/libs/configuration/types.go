package configuration

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

type Version string

func (version Version) major() (uint, error) {
	majorPart, _, _ := strings.Cut(string(version), ".")
	major, err := strconv.ParseUint(majorPart, 10, 10)
	return uint(major), err
}

func (version Version) Major() uint {
	major, _ := version.major()
	return major
}

func (version Version) minor() (uint, error) {
	_, minorPart, _ := strings.Cut(string(version), ".")
	minor, err := strconv.ParseUint(minorPart, 10, 10)
	return uint(minor), err
}

func (version Version) Minor() uint {
	minor, _ := version.minor()
	return minor
}

func MajorMinorVersion(major, minor uint) Version {
	return Version(fmt.Sprintf("%d.%d", major, minor))
}

// VersionedParseInfo 定义了如何将配置的特定版本解析为当前版本
type VersionedParseInfo struct {
	Version        Version                                // 当前版本
	ParseAs        reflect.Type                           // 当前版本配置文件，应该解析成的类型
	ConversionFunc func(interface{}) (interface{}, error) // 将当前版本的配置类型(ParseAs) 转换成最新版本的的转换函数
}

// 用于表示环境变量
type envVar struct {
	name  string
	value string
}

type envVars []envVar

func (e envVars) Len() int {
	return len(e)
}

func (e envVars) Swap(i, j int) {
	e[i], e[j] = e[j], e[i]
}
func (e envVars) Less(i, j int) bool {
	return e[i].name < e[j].name
}

// Parameters 类型定义 key-value 键值对
type Parameters map[string]interface{}

type Storage map[string]Parameters

// Type 返回存储驱动类型，如filesystem或s3
func (storage Storage) Type() string {
	var storageType []string

	// 遍历所有类型
	for k := range storage {
		switch k {
		case "maintenance":
			// 允许配置为 maintenance
		case "cache":
			// 允许配置为 caching
		case "delete":
			// 允许配置为 delete
		case "redirect":
			// 允许配置为 redirect
		case "tag":
			// 允许配置为 tag
		default:
			storageType = append(storageType, k)
		}
	}

	// 如果配置了多个 storage，报错，只允许配置一个
	if len(storageType) > 1 {
		panic("multiple storage drivers specified in configuration or environment: " + strings.Join(storageType, ", "))
	}
	if len(storageType) == 1 {
		return storageType[0]
	}
	return ""
}

func (storage Storage) Parameters() Parameters {
	return storage[storage.Type()]
}
