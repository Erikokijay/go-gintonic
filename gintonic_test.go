package gintonic

import (
	"fmt"
	"testing"

	"github.com/gin-gonic/gin"
)

type Resp struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

type Req2 struct {
	Code int    `form:"code"`
	Msg  string `form:"msg"`
}

func ping(c *gin.Context, data Resp) *Resp {
	fmt.Println(data)
	return &Resp{Code: data.Code, Msg: data.Msg + " modified"}
}

func ping2(c *gin.Context, data Req2) (int, interface{}) {
	return 200, &Resp{Code: data.Code + 1, Msg: data.Msg + " modified"}
}

func TestMain(t *testing.T) {
	gtc := NewServer(&ConfigSchema{
		SwaggerUrl:  "/docs",
		Title:       "Test",
		Description: "Desc",
		Version:     "111",
	})

	gtc.POST("/get", ping, RouteInfo{
		Tags:        []string{"Test"},
		Title:       "Route Title",
		Description: "Route Description",
	})

	gtc.GET("/test", ping2, RouteInfo{
		Tags:        []string{"Test"},
		Title:       "Route Titlttte",
		Description: "Route Description",
	}, ResultInfo{
		Code:   200,
		Output: Resp{},
	}, ResultInfo{
		Code:   401,
		Output: 0,
	})

	gtc.Run("localhost:8000")
}
