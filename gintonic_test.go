package gintonic

import (
	"testing"

	"github.com/gin-gonic/gin"
)

type Resp struct {
	Code int    `json:"code" binding:"required"`
	Msg  string `json:"msg"`
	Data []Req2 `json:"data"`
}

type Req2 struct {
	Code int    `form:"code" json:"code" binding:"required"`
	Msg  string `form:"msg" json:"msg"`
}

func ping(c *gin.Context, data Req2) *[]Resp {
	return &[]Resp{{Code: data.Code, Msg: data.Msg + " modified"}}
}

func ping2(c *gin.Context, data Req2) (int, interface{}) {
	return 200, &Resp{Code: data.Code + 1, Msg: data.Msg + " modified"}
}

func TestMain(t *testing.T) {

	eng := gin.Default()
	Config(&ConfigSchema{
		SwaggerUrl: "/docs",
	}, eng)

	router := NewRouter(eng.Group("/api"))

	router.Post("/get", ping, RouteInfo{
		Tags:        []string{"Test"},
		Title:       "Route Title",
		Description: "Route Description",
	})

	router.Get("/test", ping2, RouteInfo{
		Tags:        []string{"Test"},
		Title:       "Route Titlttte",
		Description: "Route Description",
	}, ResultInfo{
		Code:   200,
		Output: []Req2{},
	}, ResultInfo{
		Code:   401,
		Output: Req2{},
	})

	rt := NewRouter(router.Group("bbb"))
	rt.Post("/get", ping, RouteInfo{
		Tags:        []string{"Test"},
		Title:       "Route Title",
		Description: "Route Description",
	})

	GenerateSwagger(&ConfigSchema{Title: "Test"})
	eng.Run(":8080")
}
