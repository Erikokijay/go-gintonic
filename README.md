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

  	router := gnt.Group("/user")

	router.Post("/post", get, RouteInfo{
		Tags:        []string{"Test"},
		Title:       "Swagger Title",
		Description: "Swagger Description",
    	Version:      "0.2.1",
	})

  	router.Get("/get", post, 
		gintonic.ResultInfo{
			Code:   http.StatusOK, 
			Output: ExampleResponse{},
		},
		gintonic.ResultInfo{
			Code:              http.StatusInternalServerError, 
			Output:            "error",
			NeedAuthorization: true, // Simple "Authorization: Bearer" format
		}, 
		gintonic.RouteInfo{
			Tags:         []string{"Test", "First"},
			Title:        "Route Title",
			Description:  "Route Description"
		},
  	)

	gnt.Start(":8080")
}

func ping(c *gin.Context, data Req) *Resp {
	return &Resp{Code: data.Code + 1, Msg: data.Msg + " modified"} // 200 - status code
}

func ping2(c *gin.Context, data Req) (int, interface{}) { // status code, response
	return http.StatusOK, &Resp{Msg: data.Name + " modified"}
}

type Req struct {
	Code int     `json:"code" form:"code"`
	Msg  string  `json:"msg" form:"msg"`
}

func (data *Req) AutoValidate() error { // only that type of function, to automaticaly validate incoming data

	if data.Code > 1000 {
		data.Code = 1000
	} else if data.Code < 0 {
		return fmt.Errorf("invalid code")
	}

	return nil
}

type Resp struct {
	Code int     `json:"code"`
	Msg  string  `json:"msg"`
}
```
