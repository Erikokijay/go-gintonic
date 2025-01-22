#GO-GINTONIC

Module base on [![GIN](https://github.com/gin-gonic/gin)](https://github.com/gin-gonic/gin)

A basic example:

```go
package main

import (
  "net/http"

  "github.com/gin-gonic/gin"
  "github.com/eriko_kijay/go-gintonic"
)

type ExampleResponse struct {
	Msg  string `json:"msg"`
}

type ExampleRequest struct {
  Name string `json:"name"`
  Surname string `json:"surname"`
}

func main() {
  server := gintonic.NewServer()

  server.POST("/test", func(c *gin.Context, data ExampleRequest) *ExampleResponse {
	  return &ExampleResponse{Msg: data.Name + data.Surname}
  })

  server.POST("/test2", func(c *gin.Context, data ExampleRequest) (interface{}, int) {
	  return &ExampleResponse{Msg: data.Name + data.Surname}
  }, 
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
  })

  server.Run(":8000")
}


func ping(c *gin.Context, data Resp) *Resp {
	return &Resp{Code: data.Code + 1, Msg: data.Msg + " modified"}
}

func ping2(c *gin.Context, data Resp) (int, interface{}) {
	return 200, &Resp{Code: data.Code + 1, Msg: data.Msg + " modified"}
}

```