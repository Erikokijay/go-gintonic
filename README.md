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
	"fmt"
)

func main() {

	engine := gin.Default()
	gnt.Config(&gnt.ConfigSchema{
		Title:        "Test",
		Description:  "Test",
		Version:      "1.0.0",
		SwaggerUrl:   "/docs",
		SwaggerIPs:   []string{"190.0.23.222", "localhost", "::1"},
	}, engine)

  	router := gnt.Group("/user", gnt.GroupInfo{
		Title:             "GROUP",
		NeedAuthorization: true,
	})

	router.Post("/post", ping2, gnt.RouteInfo{
		Tags:        []string{"Test"},
		Title:       "Route Title",
		Description: "Route Description",
	})

  	router.Get("/get", ping, 
		gnt.ResultsInfo{
			http.StatusOK: Resp{},
			http.StatusInternalServerError: "error",
		},
		gnt.RouteInfo{
			Tags:         []string{"Test", "First"},
			Title:        "Route Title",
			Description:  "Route Description",
			NeedAuthorization: true, // Simple "Authorization: Bearer" format
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
