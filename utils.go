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

	if t.Kind() == reflect.Slice || t.Kind() == reflect.Array {
		schema["type"] = "array"
		schema["items"] = generateSchema(t.Elem())
		return schema
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
		//form := field.Tag.Get("form")

		if jsonTag != "" {
			fieldType := field.Type
			fieldKind := fieldType.Kind()
			fieldTypeStr := fieldKind.String()

			if fieldKind == reflect.Struct {
				properties[jsonTag] = generateSchema(fieldType)
			} else if fieldKind == reflect.Slice || fieldKind == reflect.Array {
				fieldTypeStr = "array"
				properties[jsonTag] = map[string]interface{}{
					"type":  fieldTypeStr,
					"items": generateSchema(fieldType.Elem()),
				}
			} else if fieldKind == reflect.Map {
				fieldTypeStr = "object"
				properties[jsonTag] = map[string]interface{}{
					"type":                 fieldTypeStr,
					"additionalProperties": generateSchema(fieldType.Elem()),
				}
			} else {
				if fieldTypeStr == "int" {
					fieldTypeStr = "integer"
				}
				if fieldTypeStr == "bool" {
					fieldTypeStr = "boolean"
				}
				if strings.Contains(fieldTypeStr, "float") {
					fieldTypeStr = "number"
				}
				properties[jsonTag] = map[string]interface{}{
					"type": fieldTypeStr,
				}
			}

			if field.Tag.Get("binding") == "required" {
				requireds = append(requireds, jsonTag)
			}
		}
	}

	schema["properties"] = properties
	schema["required"] = requireds
	return schema
}

func generateParametres(t reflect.Type, isGet bool) []Parameter {
	res := []Parameter{}

	numFields := t.NumField()

	for i := 0; i < numFields; i++ {

		field := t.Field(i)
		queryTag := field.Tag.Get("form")
		jsonTag := field.Tag.Get("json")

		if queryTag != "" && (jsonTag == "" || isGet) {

			param := Parameter{}

			fieldTypeStr := field.Type.Kind().String()

			if fieldTypeStr == "int" {
				fieldTypeStr = "integer"
			}
			if fieldTypeStr == "bool" {
				fieldTypeStr = "boolean"
			}
			if strings.Contains(fieldTypeStr, "float") {
				fieldTypeStr = "number"
			}

			param.Name = queryTag
			param.In = "query"
			param.Required = field.Tag.Get("binding") == "required"
			param.Schema = map[string]interface{}{
				"type": fieldTypeStr,
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
	if numIn < 1 {
		panic("Handler must have exactly 1 input parameters (gin.Context), recomend (gin.Context, interface{})")
	}

	if handlerType.In(0) != reflect.TypeOf((*gin.Context)(nil)) {
		panic("First parameter must be *gin.Context")
	}

	var dataType reflect.Type
	if numIn == 2 {
		dataType = handlerType.In(1)
	}
	//dataValue := reflect.New(dataType).Interface()

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

	if route.Info.Title == "" {
		route.Info.Title = handlerType.Name()
	}

	routes = append(routes, route)

	return func(c *gin.Context) {

		var inValues []reflect.Value

		if handlerType.NumIn() == 2 {
			myType := handlerType.In(1)
			myValue := reflect.New(myType).Interface()

			if err := c.ShouldBind(myValue); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}

			// Создаем слайс для входных параметров
			inValues = []reflect.Value{reflect.ValueOf(c), reflect.ValueOf(myValue).Elem()}

		} else {
			inValues = []reflect.Value{reflect.ValueOf(c)}
		}

		outValues := handlerValue.Call(inValues)

		// Проверяем, что функция возвращает один параметр
		if len(outValues) != 1 && len(outValues) != 2 {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Handler must return exactly one value"})
			return
		}

		// Возвращаем результат
		if len(outValues) == 1 {

			if outValues[0].IsNil() {
				c.Status(http.StatusBadRequest)
				return
			}

			c.JSON(http.StatusOK, outValues[0].Interface())

		} else if len(outValues) == 2 {

			if outValues[0].Kind() != reflect.Int {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Handler must return status code"})
				return
			}

			if outValues[1].IsNil() {
				c.Status(outValues[0].Interface().(int))
				return
			}

			c.JSON(outValues[0].Interface().(int), outValues[1].Interface())
		}

		c.Status(http.StatusBadRequest)
	}
}

func GenerateSwagger() {
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

		if res, ok := openAPI.Paths[route.Path]; ok {
			path = res
		}

		operation := Operation{
			Summary:     route.Info.Title,
			Description: route.Info.Description,
			Tags:        route.Info.Tags,
		}

		if route.Method != "get" && route.Method != "delete" && route.InType != nil {

			operation.RequestBody = &RequestBody{
				Required: true,
				Content: map[string]MediaType{
					"application/json": {
						Schema: Schema{"$ref": "#/components/schemas/" + route.InType.(reflect.Type).Name()},
					},
				},
			}
		}

		if route.InType != nil {
			name := ""
			if route.InType.(reflect.Type).Kind() == reflect.Slice {
				name = "Ar" + route.InType.(reflect.Type).Elem().Name()
			} else {
				name = route.InType.(reflect.Type).Name()
			}
			operation.Parameters = generateParametres(route.InType.(reflect.Type), route.Method == "get" || route.Method == "delete")
			openAPI.Components.Schemas[name] = generateSchema(route.InType.(reflect.Type))
		}

		operation.Responses = map[string]Response{}

		for code, response := range route.Responses {

			if response.(reflect.Type).Kind() == reflect.Array || response.(reflect.Type).Kind() == reflect.Struct || response.(reflect.Type).Kind() == reflect.Slice {
				name := ""
				if response.(reflect.Type).Kind() == reflect.Slice {
					name = "Ar" + response.(reflect.Type).Elem().Name()
				} else {
					name = response.(reflect.Type).Name()
				}

				operation.Responses[fmt.Sprintf("%d", code)] = Response{
					Description: name,
					Content: map[string]MediaType{
						"application/json": {
							Schema: Schema{"$ref": "#/components/schemas/" + name},
						},
					},
				}
				openAPI.Components.Schemas[name] = generateSchema(response.(reflect.Type))
			}
		}

		path[route.Method] = operation
		openAPI.Paths[route.Path] = path

		//openAPI["paths"].(map[string]interface{})[route.Path].(map[string]interface{})[route.Method].(map[string]interface{})["summary"] = route.Info
	}

	file, err := os.Create("docs/openapi.json")
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

	SaveHTML(openAPI.Info.Title)

}

func SaveHTML(title string) {
	sv := fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">

<head>
    <meta charset="UTF-8">
    <title>%s</title>
    <link rel="stylesheet" type="text/css" href="https://unpkg.com/swagger-ui-dist/swagger-ui.css" />
    <style>
        html {
            box-sizing: border-box;
            overflow: -moz-scrollbars-vertical;
            overflow-y: scroll;
        }

        *,
        *:before,
        *:after {
            box-sizing: inherit;
        }

        body {
            margin: 0;
            background: #fafafa;
        }
    </style>
</head>

<body>
    <div id="swagger-ui"></div>
    <script src="https://unpkg.com/swagger-ui-dist/swagger-ui-bundle.js"></script>
    <script src="https://unpkg.com/swagger-ui-dist/swagger-ui-standalone-preset.js"></script>
    <script>
        window.onload = function () {
            const ui = SwaggerUIBundle({
                url: "/openapi.json",
                dom_id: '#swagger-ui',
                deepLinking: true,
                presets: [
                    SwaggerUIBundle.presets.apis,
                    SwaggerUIStandalonePreset
                ],
                layout: "StandaloneLayout"
            })

            window.ui = ui

            document.getElementsByClassName("topbar")[0].style.display = "none"
        }
    </script>
</body>

</html>`, title)

	file, err := os.Create("docs/swagger.html")
	if err != nil {
		fmt.Println("Error creating swagger.html:", err)
		return
	}
	defer file.Close()

	file.WriteString(sv)

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
