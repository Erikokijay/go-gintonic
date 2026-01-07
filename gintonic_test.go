package gintonic

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

type Resp struct {
	Code  float64            `json:"code,omitempty" binding:"required"`
	Msg   string             `json:"msg"`
	Data  []Req2             `json:"data"`
	Items map[string]float64 `json:"items"`
	T     time.Time          `json:"t"`
}

type Req2 struct {
	Code int    `form:"code" json:"code,omitempty" binding:"required"`
	Msg  string `json:"msg" form:"msg"`
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
	return 200, &Resp{Code: 1231456663 / 0.00120000231231253456324512300010012, Msg: data.Msg + " modified"}
}

func TestMain(t *testing.T) {

	eng := gin.Default()

	Config(&ConfigSchema{
		SwaggerUrl: "/docs",
		SwaggerIPs: []string{"127.0.0.1", "localhost", "::1"},
		Title:      "Test",
	}, eng)

	router := Group("/api", GroupInfo{
		Title:             "sss",
		NeedAuthorization: true,
	})

	router.Post("/buy", ping2,
		RouteInfo{
			Title:             "Route Title",
			Description:       "Route Description",
			NeedAuthorization: true,
			Tags:              []string{"ss", "aa"},
		},
		ResultsInfo{
			http.StatusInternalServerError: 0,
			http.StatusOK:                  Resp{},
		},
	)
	router.Post("/", ping)

	b := Group("/bbb", GroupInfo{
		Title:             "Tbion",
		NeedAuthorization: true,
	})

	b.Get("/buy", ping2)
	b.Post("/eee", ping2)

	Start(":8080")
}
