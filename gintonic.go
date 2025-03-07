package gintonic

import (
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

type Router struct {
	*gin.RouterGroup
}

var conf *ConfigSchema

func Config(config *ConfigSchema, eng *gin.Engine) {
	if config == nil {
		conf = &ConfigSchema{
			Mode:       gin.DebugMode,
			SwaggerUrl: "/docs",
		}
	} else {
		conf = config
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
	g.GET(path, SimpleWrapper(g.BasePath()+path, handler, "GET", checkRouter(g.BasePath(), configs...)...))
}
func (g *Router) Post(path string, handler interface{}, configs ...interface{}) {
	g.POST(path, SimpleWrapper(g.BasePath()+path, handler, "POST", checkRouter(g.BasePath(), configs...)...))
}
func (g *Router) Put(path string, handler interface{}, configs ...interface{}) {
	g.PUT(path, SimpleWrapper(g.BasePath()+path, handler, "PUT", checkRouter(g.BasePath(), configs...)...))
}
func (g *Router) Delete(path string, handler interface{}, configs ...interface{}) {
	g.DELETE(path, SimpleWrapper(g.BasePath()+path, handler, "DELETE", checkRouter(g.BasePath(), configs...)...))
}

func checkRouter(path string, configs ...interface{}) []interface{} {
	cnf := configs
	if path != "" && path != "/" {
		for i := range cnf {
			if routeInfo, ok := cnf[i].(RouteInfo); ok {
				routeInfo.Tags = append(routeInfo.Tags, strings.Split(path, "/")[0])
				cnf[i] = routeInfo
			}
		}
	}

	return cnf
}
