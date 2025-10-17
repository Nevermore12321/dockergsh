package v1

import (
	"github.com/Nevermore12321/dockergsh/external/registry/registry/api/errcode"
	"net/http"
	"regexp"
)

const (
	RouteNameBase    = "base"
	RouteNameCatalog = "catalog"
)

// RouteDescriptor 定义路由
type RouteDescriptor struct {
	Name        string             // 路由名称
	Path        string             // 路由路径
	Entity      string             // 路由针对的资源目标
	Description string             // 描述信息
	Methods     []MethodDescriptor // http 各种请求方法，包括请求和响应格式
}

type MethodDescriptor struct {
	Method      string              // http 方法，PUT、GET...
	Description string              // 描述信息
	Requests    []RequestDescriptor // 规定了请求如何定义
}

type RequestDescriptor struct {
	Name            string                // 为该请求提供一个简单的名字
	Description     string                // 描述信息
	Headers         []ParameterDescriptor // 请求头
	PathParameters  []ParameterDescriptor // 请求路径中的参数
	QueryParameters []ParameterDescriptor // 查询参数
	Body            BodyDescriptor        // 请求体
	Successes       []ResponseDescriptor  // 成功的响应
	Failures        []ResponseDescriptor  // 失败的响应
}

type ResponseDescriptor struct {
	Name        string                // 为该响应提供一个简单的名字
	Description string                // 描述信息
	StatusCode  int                   // 响应状态码
	Headers     []ParameterDescriptor // 响应头
	Fields      []ParameterDescriptor // 响应中可能包含的字段
	ErrorCodes  []errcode.ErrorCode   // 返回的自定义错误码
	Body        BodyDescriptor        // 响应结构
}

type BodyDescriptor struct {
	ContentType string // 定义 Content-type，json or form?
	Format      string
}

type ParameterDescriptor struct {
	Name        string
	Type        string         // Type指定参数的类型，如字符串、整数等。
	Description string         // 描述信息
	Required    bool           // 是否是必须的
	Format      string         // 指定参数接受的字符串格式
	Regexp      *regexp.Regexp // 参数的正则匹配
	Examples    []string       // 示例信息
}

// 定义一些常用的 参数、头 等信息
var (
	hostHeader = ParameterDescriptor{
		Name:        "Host",
		Type:        "string",
		Description: "Standard HTTP Host Header. Should be set to the registry host.",
		Format:      "<registry host>",
		Examples:    []string{"registry-1.docker.io"},
	}
	authHeader = ParameterDescriptor{
		Name:        "Authorization",
		Type:        "string",
		Description: "An RFC7235 compliant authorization header.",
		Format:      "<scheme> <token>",
		Examples:    []string{"Bearer dGhpcyBpcyBhIGZha2UgYmVhcmVyIHRva2VuIQ=="},
	}
	authChallengeHeader = ParameterDescriptor{
		Name:        "WWW-Authenticate",
		Type:        "string",
		Description: "An RFC7235 compliant authentication challenge header.",
		Format:      `<scheme> realm="<realm>", ..."`,
		Examples: []string{
			`Bearer realm="https://auth.docker.com/", service="registry.docker.com", scopes="repository:library/ubuntu:pull"`,
		},
	}

	unauthorizedResponseDescriptor = ResponseDescriptor{
		Name:        "Authentication Required",
		StatusCode:  http.StatusUnauthorized,
		Description: "The client is not authenticated.",
		Headers: []ParameterDescriptor{
			authChallengeHeader,
			{
				Name:        "Content-Length",
				Type:        "integer",
				Description: "Length of the JSON response body.",
				Format:      "<length>",
			},
		},
		Body: BodyDescriptor{
			ContentType: "application/json",
			Format:      errorsBody,
		},
		ErrorCodes: []errcode.ErrorCode{
			errcode.ErrorCodeUnauthorized,
		},
	}

	tooManyRequestsDescriptor = ResponseDescriptor{
		Name:        "Too Many Requests",
		StatusCode:  http.StatusTooManyRequests,
		Description: "The client made too many requests within a time interval.",
		Headers: []ParameterDescriptor{
			{
				Name:        "Content-Length",
				Type:        "integer",
				Description: "Length of the JSON response body.",
				Format:      "<length>",
			},
		},
		Body: BodyDescriptor{
			ContentType: "application/json",
			Format:      errorsBody,
		},
		ErrorCodes: []errcode.ErrorCode{
			errcode.ErrorCodeTooManyRequests,
		},
	}
)
