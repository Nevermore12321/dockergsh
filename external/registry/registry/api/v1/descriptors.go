package v1

import "net/http"

var routeDescriptorsMap map[string]RouteDescriptor

func init() {
	routeDescriptorsMap = make(map[string]RouteDescriptor, len(routeDescriptors))
	for _, descriptor := range routeDescriptors {
		routeDescriptorsMap[descriptor.Name] = descriptor
	}
}

// 定义所有路由信息
var routeDescriptors = []RouteDescriptor{
	{
		Name:        RouteNameBase,
		Path:        "/v1/",
		Entity:      "Base",
		Description: `Base V1 API route. Typically, this can be used for lightweight version checks and to validate registry authentication.`,
		Methods: []MethodDescriptor{
			{
				Method:      http.MethodGet,
				Description: "Check that the endpoint implements Docker Registry API V2.",
				Requests: []RequestDescriptor{
					{
						Headers: []ParameterDescriptor{
							hostHeader,
							authHeader,
						},
						Successes: []ResponseDescriptor{
							{
								Description: "The API implements V1 protocol and is accessible.",
								StatusCode:  http.StatusOK,
							},
						},
						Failures: []ResponseDescriptor{
							{
								Description: "The registry does not implement the V1 API.",
								StatusCode:  http.StatusNotFound,
							},
							unauthorizedResponseDescriptor,
							tooManyRequestsDescriptor,
						},
					},
				},
			},
		},
	},
}
