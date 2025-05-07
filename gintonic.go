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

	conf.engine = eng

	_, err := os.Stat("./docs")
	if err != nil {
		if os.IsNotExist(err) {
			os.Mkdir("./docs", os.FileMode(0777))
		}
	}

	eng.StaticFile(conf.SwaggerUrl, "./docs/swagger.html")
	eng.StaticFile(conf.SwaggerUrl+"/openapi.json", "./docs/openapi.json")
}

func NewRouter(eng *gin.RouterGroup) *Router {
	return &Router{eng}
}

func Group(path string) *Router {
	return &Router{conf.engine.Group(path)}
}

func (r *Router) SubGroup(path string) *Router {
	return &Router{r.Group(path)}
}

// Use RouteInfo{} struct for describe router, and use ResultInfo{} to describe responses
func (g *Router) Get(path string, handler interface{}, configs ...interface{}) {
	g.GET(path, simpleWrapper(g.BasePath()+path, handler, "GET", checkRouter(g.BasePath(), configs...)...))
}

// Use RouteInfo{} struct for describe router, and use ResultInfo{} to describe responses
func (g *Router) Post(path string, handler interface{}, configs ...interface{}) {
	g.POST(path, simpleWrapper(g.BasePath()+path, handler, "POST", checkRouter(g.BasePath(), configs...)...))
}

// Use RouteInfo{} struct for describe router, and use ResultInfo{} to describe responses
func (g *Router) Put(path string, handler interface{}, configs ...interface{}) {
	g.PUT(path, simpleWrapper(g.BasePath()+path, handler, "PUT", checkRouter(g.BasePath(), configs...)...))
}

// Use RouteInfo{} struct for describe router, and use ResultInfo{} to describe responses
func (g *Router) Delete(path string, handler interface{}, configs ...interface{}) {
	g.DELETE(path, simpleWrapper(g.BasePath()+path, handler, "DELETE", checkRouter(g.BasePath(), configs...)...))
}

func checkRouter(path string, configs ...interface{}) []interface{} {

	haveTags := false
	cnf := configs
	if path != "" && path != "/" {
		if len(strings.Split(path, "/")) > 1 {

			routerName := strings.Split(path, "/")[1]
			routerName = strings.ToUpper(routerName[:1]) + routerName[1:]

			for i := range cnf {

				if routeInfo, ok := cnf[i].(RouteInfo); ok {

					routeInfo.Tags = append(routeInfo.Tags, routerName)
					cnf[i] = routeInfo
					haveTags = true
					break
				}
			}

			if !haveTags {
				cnf = append(cnf, RouteInfo{Tags: []string{routerName}})
			}
		}
	}

	return cnf
}
