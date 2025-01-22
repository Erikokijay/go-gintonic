package ginfastapi

type ConfigSchema struct {
	SwaggerUrl  string `json:"swaggerUrl"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Version     string `json:"version"`
	Mode        string `json:"mode"`
}

type ResultInfo struct {
	Code   int         `json:"code"`
	Output interface{} `json:"output"`
}

type RouteInfo struct {
	Tags        []string `json:"tags"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
}

type OpenAPI struct {
	OpenAPI    string     `json:"openapi"`
	Info       Info       `json:"info"`
	Paths      Paths      `json:"paths"`
	Components Components `json:"components"`
}

type Info struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Version     string `json:"version"`
}

type Paths map[string]PathItem

type PathItem map[string]Operation

type Operation struct {
	Tags        []string              `json:"tags"`
	Summary     string                `json:"summary"`
	Description string                `json:"description"`
	RequestBody *RequestBody          `json:"requestBody,omitempty"`
	Parameters  []Parameter           `json:"parameters,omitempty"`
	Responses   map[string]Response   `json:"responses"`
	Security    []SecurityRequirement `json:"security"`
}

type Parameter struct {
	Name     string `json:"name"`
	In       string `json:"in"`
	Required bool   `json:"required"`
	Schema   Schema `json:"schema"`
}

type RequestBody struct {
	Required bool                 `json:"required"`
	Content  map[string]MediaType `json:"content"`
}

type MediaType struct {
	Schema Schema `json:"schema"`
}

type Schema map[string]interface{}

type Response struct {
	Description string               `json:"description"`
	Content     map[string]MediaType `json:"content"`
}

type Components struct {
	Schemas         map[string]Schema         `json:"schemas"`
	SecuritySchemes map[string]SecurityScheme `json:"securitySchemes"`
}

type SecurityScheme struct {
	Type         string `json:"type"`
	Scheme       string `json:"scheme"`
	BearerFormat string `json:"bearerFormat,omitempty"`
}

type SecurityRequirement map[string][]string
