package ginfastapi

import "github.com/gin-gonic/gin"

type Resp struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

func ping(c *gin.Context, data Resp) *Resp {
	return &Resp{Code: data.Code + 1, Msg: data.Msg + " modified"}
}

func ping2(c *gin.Context, data Resp) (int, interface{}) {
	return 200, &Resp{Code: data.Code + 1, Msg: data.Msg + " modified"}
}

func main() {
	gtc := NewServer(&ConfigSchema{})

	gtc.GET("/get", ping)

	gtc.Run("localhost:8000")
}
