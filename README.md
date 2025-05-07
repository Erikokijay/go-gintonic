# GO-GINTONIC

Module base on GIN. Automatize data validating and swagger generating

Installation:

```sh
go get github.com/Erikokijay/go-gintonic
```

### EXAMPLE

```go
package main

import (
	gnt "github.com/Erikokijay/go-gintonic"
	"github.com/gin-gonic/gin"
	"net/http"
)

func main() {

	engine := gin.Default()
	gnt.Config(&gnt.ConfigSchema{
		Title: "Test",
		SwaggerUrl: "/docs",
	}, engine)

  	router := gnt.NewRouter(engine.Group("/user"))

	router.Post("/get", ping, RouteInfo{
		Tags:        []string{"Test"},
		Title:       "Swagger Title",
		Description: "Swagger Description",
    	Version:      "0.2.1",
	})

  	router.Get("/test", ping2, 
		gintonic.ResultInfo{
			Code:   http.StatusOK, 
			Output: ExampleResponse{},
		},
		gintonic.ResultInfo{
			Code:   500, 
			Output: "error",
		}, 
		gintonic.RouteInfo{
			Tags:         []string{"Test", "First"},
			Title:        "Route Title",
			Description:  "Route Description"
		},
  	)

	gnt.GenerateSwagger()
	engine.Run(":8080")
}

func ping(c *gin.Context, data Resp) *Resp {
	return &Resp{Code: data.Code + 1, Msg: data.Msg + " modified"} // 200 - status code
}

func ping2(c *gin.Context, data ExampleRequest) (int, interface{}) { // status code, response
	return http.StatusOK, &ExampleResponse{Msg: data.Name + " modified"}
}
```
