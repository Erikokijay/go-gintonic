package gintonic

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
)

type Resp struct {
	Code  int                `json:"code,omitempty" binding:"required"`
	Msg   string             `json:"msg"`
	Data  []Req2             `json:"data"`
	Items map[string]float64 `json:"items"`
}

type Req2 struct {
	Code int    `form:"code" json:"code,omitempty" binding:"required"`
	Msg  string `json:"msg"`
	Bb   bool   `form:"bb" json:"bb"`
}

func (r *Req2) AutoValidate() error {

	if r.Code > 100 {
		return fmt.Errorf("CODE")
	}

	return nil
}

func ping(c *gin.Context) *[]Resp {
	return &[]Resp{}
}

func ping2(c *gin.Context, data Req2) (int, interface{}) {
	return 200, &Resp{Code: data.Code + 1, Msg: data.Msg + " modified"}
}

func TestMain(t *testing.T) {

	eng := gin.Default()

	Config(&ConfigSchema{
		SwaggerUrl: "/docs",
		Title:      "Test",
	}, eng)

	router := Group("/api")

	router.Post("/buy", ping2,
		RouteInfo{
			Title:             "Route Title",
			Description:       "Route Description",
			NeedAuthorization: true,
		},
		ResultsInfo{
			http.StatusInternalServerError: 0,
			http.StatusOK:                  "ok",
		},
	)
	router.Post("/", ping)

	b := Group("/bbb", GroupInfo{
		Title:             "Tbion",
		NeedAuthorization: true,
	})

	b.Get("/buy", ping2)
	b.Get("/eee", ping2)

	Start(":8080")
}
