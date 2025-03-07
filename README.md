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
	"github.com/joho/godotenv"
)

func main() {
	gin.SetMode("release")
	eng := gin.Default()
	gnt.Config(&gnt.ConfigSchema{
		Title: "Test",
		SwaggerUrl: "/docs",
	}, eng)

  	router := gnt.NewRouter(eng.Group("/api"))

	router.Post("/get", ping, RouteInfo{
		Tags:        []string{"Test"},
		Title:       "Route Title",
		Description: "Route Description",
	})

  	router.Get("/test", ping2, 
		gintonic.ResultInfo{
			Code: 200, 
			Output: ExampleResponse{},
		},
		gintonic.ResultInfo{
			Code: 500, 
			Output: "error",
		}, 
		gintonic.RouteInfo{
			Tags: []string{"Test", "First"},
			Title: "Route Title",
			Description: "Route Description"
		},
  	)

	gnt.GenerateSwagger(&gnt.ConfigSchema{
		Title:   "Test",
		Version: "1.2.1",
		Mode:    "release",
	})
	eng.Run(":8080")
}

func ping(c *gin.Context, data Resp) *Resp {
	return &Resp{Code: data.Code + 1, Msg: data.Msg + " modified"} // 200 - status code
}

func ping2(c *gin.Context, data ExampleRequest) (int, interface{}) {
	return http.StatusOK, &ExampleResponse{Msg: data.Name + " modified"}
}
```
