package gintonic

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"reflect"
	"strings"

	"github.com/gin-gonic/gin"
)

func generateSchema(t reflect.Type) map[string]interface{} {
	schema := map[string]interface{}{
		"type": "object",
	}

	if t.Kind() != reflect.Struct {
		return nil
	}

	numFields := t.NumField()
	properties := make(map[string]interface{}, numFields)
	requireds := []string{}

	for i := 0; i < numFields; i++ {
		field := t.Field(i)
		jsonTag := field.Tag.Get("json")

		if jsonTag != "" {
			fieldType := field.Type.Kind().String()

			if fieldType == "int" {
				fieldType = "integer"
			}
			if strings.Contains(fieldType, "float") {
				fieldType = "number"
			}

			properties[jsonTag] = map[string]interface{}{
				"type": fieldType,
			}

			if field.Tag.Get("requeuired") == "true" {
				requireds = append(requireds, jsonTag)
			}
		}
	}

	schema["properties"] = properties
	schema["required"] = requireds
	return schema
}

func generateParametres(t reflect.Type) []Parameter {
	res := []Parameter{}

	numFields := t.NumField()

	for i := 0; i < numFields; i++ {

		field := t.Field(i)
		queryTag := field.Tag.Get("form")

		if queryTag != "" {

			param := Parameter{}

			fieldType := field.Type.Kind().String()

			param.Name = queryTag
			param.In = "query"
			param.Required = field.Tag.Get("requeuired") == "true"
			param.Schema = map[string]interface{}{
				"type": fieldType,
			}

			res = append(res, param)
		}
	}

	return res
}

type ApiRoute struct {
	Path      string
	InType    interface{}
	Method    string
	Info      RouteInfo
	Responses map[int]interface{}
}

var routes []ApiRoute = []ApiRoute{}

// SimpleWrapper - create simple wrapper for gin handler, to validate data and create swagger.
// path - route full path, handler - handler function, method - http method (GET, PUT, DELETE, POST), configs - ...RouteInfo, ...ResultInfo
func SimpleWrapper(path string, handler interface{}, method string, configs ...interface{}) gin.HandlerFunc {

	handlerType := reflect.TypeOf(handler)
	handlerValue := reflect.ValueOf(handler)

	if handlerType.Kind() != reflect.Func {
		panic("Handler is not a function")
	}

	numIn := handlerType.NumIn()
	if numIn != 2 {
		panic("Handler must have exactly 2 input parameters (gin.Context, interface{})")
	}

	if handlerType.In(0) != reflect.TypeOf((*gin.Context)(nil)) {
		panic("First parameter must be *gin.Context")
	}

	dataType := handlerType.In(1)
	dataValue := reflect.New(dataType).Interface()

	// Добавляем маршрут в список маршрутов
	route := ApiRoute{
		Path:      path,
		InType:    dataType,
		Method:    strings.ToLower(method),
		Responses: map[int]interface{}{},
	}

	for _, conf := range configs {
		if reflect.TypeOf(conf).Name() == "ResultInfo" {
			res := conf.(ResultInfo)
			route.Responses[res.Code] = reflect.TypeOf(res.Output)
		} else if reflect.TypeOf(conf).Name() == "RouteInfo" {
			route.Info = conf.(RouteInfo)
		}
	}

	if handlerType.NumOut() == 1 {
		route.Responses[200] = handlerType.Out(0).Elem()
	}

	routes = append(routes, route)

	return func(c *gin.Context) {
		if err := c.ShouldBind(dataValue); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Создаем слайс для входных параметров
		inValues := []reflect.Value{reflect.ValueOf(c), reflect.ValueOf(dataValue).Elem()}

		// Вызываем функцию с входными параметрами
		outValues := handlerValue.Call(inValues)

		// Проверяем, что функция возвращает один параметр
		if len(outValues) != 1 {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Handler must return exactly one value"})
			return
		}

		// Возвращаем результат
		c.JSON(http.StatusOK, outValues[0].Interface())
	}
}

func GenerateSwagger(conf *ConfigSchema) {
	openAPI := OpenAPI{
		OpenAPI: "3.1.0",
		Info: Info{
			Title:       conf.Title,
			Description: conf.Description,
			Version:     conf.Version,
		},
		Paths:      Paths{},
		Components: Components{Schemas: map[string]Schema{}},
	}

	for _, route := range routes {

		path := PathItem{}

		operation := Operation{
			Summary:     route.Info.Title,
			Description: route.Info.Description,
			Tags:        route.Info.Tags,
		}

		if route.Method != "get" && route.Method != "delete" {

			operation.RequestBody = &RequestBody{
				Required: true,
				Content: map[string]MediaType{
					"application/json": {
						Schema: Schema{"$ref": "#/components/schemas/" + route.InType.(reflect.Type).Name()},
					},
				},
			}
		}

		operation.Parameters = generateParametres(route.InType.(reflect.Type))
		openAPI.Components.Schemas[route.InType.(reflect.Type).Name()] = generateSchema(route.InType.(reflect.Type))

		operation.Responses = map[string]Response{}

		for code, response := range route.Responses {

			operation.Responses[fmt.Sprintf("%d", code)] = Response{
				Description: response.(reflect.Type).Name(),
				Content: map[string]MediaType{
					"application/json": {
						Schema: Schema{"$ref": "#/components/schemas/" + response.(reflect.Type).Name()},
					},
				},
			}
			openAPI.Components.Schemas[response.(reflect.Type).Name()] = generateSchema(response.(reflect.Type))
		}

		path[route.Method] = operation
		openAPI.Paths[route.Path] = path

		//openAPI["paths"].(map[string]interface{})[route.Path].(map[string]interface{})[route.Method].(map[string]interface{})["summary"] = route.Info
	}

	file, err := os.Create("static/openapi.json")
	if err != nil {
		fmt.Println("Error creating openapi.json:", err)
		return
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(openAPI); err != nil {
		fmt.Println("Error encoding openapi.json:", err)
	}
}

/*
func generateOpenAPI(path string, handlerType reflect.Type, inType reflect.Type, outType reflect.Type) {
	openAPI := map[string]interface{}{
		"openapi": "3.1.0",
		"info": map[string]interface{}{
			"title":       "API Documentation",
			"description": "API for managing users",
			"version":     "1.0.0",
		},
		"paths": map[string]interface{}{
			path: map[string]interface{}{
				"post": map[string]interface{}{
					"summary":     "Handle POST request",
					"description": "Handle POST request to " + path,
					"requestBody": map[string]interface{}{
						"required": true,
						"content": map[string]interface{}{
							"application/json": map[string]interface{}{
								"schema": map[string]interface{}{
									"$ref": "#/components/schemas/" + inType.Name(),
								},
							},
						},
					},
					"responses": map[string]interface{}{
						"200": map[string]interface{}{
							"description": "Successful response",
							"content": map[string]interface{}{
								"application/json": map[string]interface{}{
									"schema": map[string]interface{}{
										"$ref": "#/components/schemas/" + outType.Name(),
									},
								},
							},
						},
					},
				},
			},
		},
		"components": map[string]interface{}{
			"schemas": map[string]interface{}{
				inType.Name():  generateSchema(inType),
				outType.Name(): generateSchema(outType),
			},
		},
	}

	// Записываем OpenAPI спецификацию в файл
	file, err := os.Create("static/openapi.json")
	if err != nil {
		fmt.Println("Error creating openapi.json:", err)
		return
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(openAPI); err != nil {
		fmt.Println("Error encoding openapi.json:", err)
	}
}*/
