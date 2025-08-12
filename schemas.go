package gintonic

import "github.com/gin-gonic/gin"

type ConfigSchema struct {
	SwaggerUrl  string `json:"swaggerUrl"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Version     string `json:"version"`
	Mode        string `json:"mode"`
	engine      *gin.Engine
}

type ResultInfo struct {
	Code   int         `json:"code"`
	Output interface{} `json:"output"`
}

type RouteInfo struct {
	Tags              []string `json:"tags"`
	Title             string   `json:"title"`
	Description       string   `json:"description"`
	NeedAuthorization bool     `json:"need_authorization"`
}

type GroupInfo struct {
	Title             string `json:"title"`
	NeedAuthorization bool   `json:"need_authorization"`
}

type openAPI struct {
	OpenAPI    string     `json:"openapi"`
	Info       info       `json:"info"`
	Paths      paths      `json:"paths"`
	Components components `json:"components"`
}

type info struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Version     string `json:"version"`
}

type paths map[string]pathItem

type pathItem map[string]operation

type operation struct {
	Tags        []string              `json:"tags"`
	Summary     string                `json:"summary"`
	Description string                `json:"description"`
	RequestBody *requestBody          `json:"requestBody,omitempty"`
	Parameters  []parameter           `json:"parameters,omitempty"`
	Responses   map[string]response   `json:"responses"`
	Security    []securityRequirement `json:"security"`
}

type parameter struct {
	Name     string `json:"name"`
	In       string `json:"in"`
	Required bool   `json:"required"`
	Schema   schema `json:"schema"`
}

type requestBody struct {
	Required bool                 `json:"required"`
	Content  map[string]mediaType `json:"content"`
}

type mediaType struct {
	Schema schema `json:"schema"`
}

type schema map[string]interface{}

type response struct {
	Description string               `json:"description"`
	Content     map[string]mediaType `json:"content"`
}

type components struct {
	Schemas         map[string]interface{}    `json:"schemas"`
	SecuritySchemes map[string]securityScheme `json:"securitySchemes"`
}

type securityScheme struct {
	Type         string `json:"type"`
	Scheme       string `json:"scheme"`
	BearerFormat string `json:"bearerFormat,omitempty"`
	Description  string `json:"description,omitempty"`
}

type SecurityType string

const (
	Bearer SecurityType = "bearer"
)

type securityRequirement map[string][]string
