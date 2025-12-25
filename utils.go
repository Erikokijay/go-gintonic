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

func generateSchema(t reflect.Type) interface{} {

	schema := map[string]interface{}{
		"type": "object",
	}

	if strings.Contains(t.Kind().String(), "int") {
		schema["type"] = "integer"
		return schema
	} else if t.Kind().String() == "bool" {
		schema["type"] = "boolean"
		return schema
	} else if strings.Contains(t.Kind().String(), "float") {
		schema["type"] = "number"
		return schema
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

		if spl := strings.Split(jsonTag, ","); len(spl) > 0 {
			jsonTag = spl[0]
		}
		//form := field.Tag.Get("form")

		if jsonTag != "" {
			fieldType := field.Type
			fieldKind := fieldType.Kind()
			fieldTypeStr := strings.ReplaceAll(fieldKind.String(), ",omitempty", "")

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
				if strings.Contains(fieldTypeStr, "int") {
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

func generateParametres(t reflect.Type, isGet bool, path string) []parameter {
	res := []parameter{}

	numFields := t.NumField()

	for i := range numFields {

		field := t.Field(i)
		queryTag := field.Tag.Get("form")
		if spl := strings.Split(queryTag, ","); len(spl) > 0 {
			queryTag = spl[0]
		}

		jsonTag := field.Tag.Get("json")
		if spl := strings.Split(jsonTag, ","); len(spl) > 0 {
			jsonTag = spl[0]
		}

		if isGet && queryTag == "" && jsonTag != "" {
			fmt.Println("\033[33m⚠️  WARNING! In GET ("+path+") method you have json, and not have form parametres. Json Tag:", jsonTag, "\033[0m")
		}

		if queryTag != "" && (jsonTag == "" || isGet) {

			param := parameter{}

			fieldTypeStr := strings.ReplaceAll(field.Type.Kind().String(), ",omitempty", "")

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

type apiRoute struct {
	Path      string
	InType    interface{}
	Method    string
	Info      RouteInfo
	Responses map[int]interface{}
}

var routes []apiRoute = []apiRoute{}

// SimpleWrapper - create simple wrapper for gin handler, to validate data and create swagger.
// path - route full path, handler - handler function, method - http method (GET, PUT, DELETE, POST), configs - ...RouteInfo, ...ResultInfo
func simpleWrapper(path string, handler interface{}, method string, configs ...interface{}) gin.HandlerFunc {

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

	route := apiRoute{
		Path:      path,
		InType:    dataType,
		Method:    strings.ToLower(method),
		Responses: map[int]interface{}{},
	}

	haveOut := false
	for _, conf := range configs {
		if reflect.TypeOf(conf).Name() == "ResultInfo" {
			res := conf.(ResultInfo)
			route.Responses[res.Code] = reflect.TypeOf(res.Output)
			haveOut = true

		} else if reflect.TypeOf(conf).Name() == "ResultsInfo" {
			res := conf.(ResultsInfo)
			for code, out := range res {
				route.Responses[code] = reflect.TypeOf(out)
				haveOut = true
			}

		} else if reflect.TypeOf(conf).Name() == "RouteInfo" {
			route.Info = conf.(RouteInfo)
		}
	}

	if handlerType.NumOut() == 1 {
		route.Responses[200] = handlerType.Out(0).Elem()
	} else if !haveOut {
		route.Responses[200] = reflect.TypeOf("")
	}

	if route.Info.Title == "" {
		route.Info.Title = handlerType.Name()
	}

	routes = append(routes, route)

	return func(c *gin.Context) {

		var inValues []reflect.Value

		if handlerType.NumIn() == 2 {
			myType := handlerType.In(1)
			isPointer := myType.Kind() == reflect.Ptr

			var myValue reflect.Value = reflect.New(myType)
			if isPointer {
				myValue = reflect.New(myType.Elem())
			}

			//fmt.Println(myValue.Type().Kind(), isPointer, myValue.Interface())
			if err := c.ShouldBind(myValue.Interface()); err != nil {
				//fmt.Println(err, "err1")
				c.JSON(http.StatusBadRequest, SimpleResponse{Ok: false, Error: err.Error()})
				return
			}

			v := myValue
			if isPointer {
				v = v.Addr()
			}

			if hasAutoValidate(v) {
				//fmt.Println("CALL")
				err := callAutoValidate(v)
				if err != nil {
					//fmt.Println(err, "err2")
					c.JSON(http.StatusBadRequest, SimpleResponse{Ok: false, Error: err.Error()})
					return
				}
			}

			inValues = []reflect.Value{reflect.ValueOf(c), reflect.ValueOf(myValue.Interface()).Elem()}

		} else {
			inValues = []reflect.Value{reflect.ValueOf(c)}
		}

		outValues := handlerValue.Call(inValues)

		if len(outValues) != 1 && len(outValues) != 2 {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Handler must return exactly one value"})
			return
		}

		if len(outValues) == 1 {

			if outValues[0].IsNil() {
				c.Status(http.StatusBadRequest)
				return
			}

			c.JSON(http.StatusOK, outValues[0].Interface())
			return

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
			return
		}

		c.Status(http.StatusBadRequest)
	}
}
func hasAutoValidate(v reflect.Value) bool {
	method := v.MethodByName("AutoValidate")
	return method.IsValid()
}

func callAutoValidate(v reflect.Value) error {
	method := v.MethodByName("AutoValidate")

	results := method.Call(nil)

	if len(results) == 1 {

		if results[0].IsNil() {
			return nil
		}
		if results[0].Kind() == reflect.Interface {
			if err, ok := results[0].Interface().(error); ok {
				return err
			}
		}
	}

	return nil
}

var haveAuth bool = false

func GenerateSwagger() {

	openAPI := openAPI{
		OpenAPI: "3.1.1",
		Info: info{
			Title:       conf.Title,
			Description: conf.Description,
			Version:     conf.Version,
		},
		Paths:      paths{},
		Components: components{Schemas: map[string]interface{}{}},
	}

	for _, route := range routes {

		if route.Info.NeedAuthorization {
			haveAuth = true
		}

		path := pathItem{}

		if res, ok := openAPI.Paths[route.Path]; ok {
			path = res
		}

		operation := operation{
			Summary:     route.Info.Title,
			Description: route.Info.Description,
			Tags:        route.Info.Tags,
		}

		if route.Method != "get" && route.Method != "delete" && route.InType != nil {

			operation.RequestBody = &requestBody{
				Required: true,
				Content: map[string]mediaType{
					"application/json": {
						Schema: schema{"$ref": "#/components/schemas/" + route.InType.(reflect.Type).Name()},
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
			operation.Parameters = generateParametres(route.InType.(reflect.Type), route.Method == "get" || route.Method == "delete", route.Path)
			openAPI.Components.Schemas[name] = generateSchema(route.InType.(reflect.Type))
		}

		operation.Responses = map[string]response{}

		for code, resp := range route.Responses {

			name := ""
			if resp.(reflect.Type).Kind() == reflect.Slice {
				name = "Ar" + resp.(reflect.Type).Elem().Name()
			} else {
				name = resp.(reflect.Type).Name()
			}

			operation.Responses[fmt.Sprintf("%d", code)] = response{
				Description: name,
				Content: map[string]mediaType{
					"application/json": {
						Schema: schema{"$ref": "#/components/schemas/" + name},
					},
				},
			}
			openAPI.Components.Schemas[name] = generateSchema(resp.(reflect.Type))

		}

		if route.Info.NeedAuthorization {
			operation.Security = []securityRequirement{
				{
					"BearerAuth": []string{},
				},
			}
		}

		path[route.Method] = operation
		openAPI.Paths[route.Path] = path

		//openAPI["paths"].(map[string]interface{})[route.Path].(map[string]interface{})[route.Method].(map[string]interface{})["summary"] = route.Info
	}

	if haveAuth {
		openAPI.Components.SecuritySchemes = map[string]securityScheme{
			"BearerAuth": {
				Type:         "http",
				Scheme:       "bearer",
				BearerFormat: "JWT",
				Description:  "JWT Authorization header using the Bearer scheme. Example: \"Authorization: Bearer {token}\"",
			},
		}
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

	authInerceptor := ""
	authBtn := ""

	if haveAuth {
		authInerceptor = `,
		requestInterceptor: (req) => {
			const token = localStorage.getItem('swagger_token');
			if (token) {
				req.headers.Authorization = 'Bearer ' + token;
			}
			return req;
		}`

		authBtn = `const btn = document.createElement('button');
		btn.innerHTML = 'Set JWT Token';
		btn.style.position = 'fixed';
		btn.style.top = '10px';
		btn.style.right = '10px';
		btn.style.zIndex = '1000';
		btn.onclick = function() {
			const token = prompt('Enter JWT Token:');
			if (token) {
				console.log(token)
				localStorage.setItem('swagger_token', token);
				//window.location.reload();
			}
		};
		document.body.appendChild(btn);`
	}

	authBtn = ""
	authInerceptor = ""

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
                url: "%s/openapi.json",
                dom_id: '#swagger-ui',
                deepLinking: true,
                presets: [
                    SwaggerUIBundle.presets.apis,
                    SwaggerUIStandalonePreset
                ],
                layout: "StandaloneLayout"%s
            })

            window.ui = ui

			%s
        }
    </script>
</body>

</html>`, title, conf.SwaggerUrl, authInerceptor, authBtn)

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
