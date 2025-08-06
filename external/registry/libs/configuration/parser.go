package configuration

// Parser 解析器可以用来解析配置文件和环境变量的定义版本，到一个统一的输出结构
type Parser struct {
	prefix  string
	mapping map[Version]VersionedParseInfo
	env     envVars
}

// NewParser 初始化 Parser 对象，给定一个环境变量前缀 prefix，以及一个特定版本的 parseInfo
func NewParser(prefix string, parseInfos []VersionedParseInfo) *Parser {

}
