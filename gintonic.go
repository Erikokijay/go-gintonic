package gintonic

import (
	"os"

	"github.com/gin-gonic/gin"
)

type Router struct {
	*gin.RouterGroup
}

func Config(conf *ConfigSchema, eng *gin.Engine) {
	if conf == nil {
		conf = &ConfigSchema{
			Mode:       gin.DebugMode,
			SwaggerUrl: "/docs",
		}
	}

	_, err := os.Stat("./docs")
	if err != nil {
		if os.IsNotExist(err) {
			os.Mkdir("./docs", os.FileMode(0777))
		}
	}

	eng.StaticFile(conf.SwaggerUrl, "./docs/swagger.html")
	eng.StaticFile("/openapi.json", "./docs/openapi.json")
}

func NewRouter(eng *gin.RouterGroup) *Router {
	return &Router{eng}
}

func (g *Router) Get(path string, handler interface{}, configs ...interface{}) {
	g.GET(path, SimpleWrapper(g.BasePath()+path, handler, "GET", configs...))
}
func (g *Router) Post(path string, handler interface{}, configs ...interface{}) {
	g.POST(path, SimpleWrapper(g.BasePath()+path, handler, "POST", configs...))
}
func (g *Router) Put(path string, handler interface{}, configs ...interface{}) {
	g.PUT(path, SimpleWrapper(g.BasePath()+path, handler, "PUT", configs...))
}
func (g *Router) Delete(path string, handler interface{}, configs ...interface{}) {
	g.DELETE(path, SimpleWrapper(g.BasePath()+path, handler, "DELETE", configs...))
}
