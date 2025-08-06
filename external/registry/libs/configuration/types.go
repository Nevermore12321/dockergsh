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
