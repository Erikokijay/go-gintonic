package gintonic

import (
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

type Router struct {
	*gin.RouterGroup
	Info GroupInfo
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

func Start(host string) {
	GenerateSwagger()
	conf.engine.Run(host)
}

func NewRouter(eng *gin.RouterGroup) *Router {
	return &Router{eng, GroupInfo{}}
}

func Group(path string, info ...GroupInfo) *Router {
	gr := GroupInfo{}

	for i := range info {
		gr = info[i]
		break
	}

	return &Router{conf.engine.Group(path), gr}
}

func (r *Router) SubGroup(path string, info ...RouteInfo) *Router {
	return &Router{r.Group(path), GroupInfo{}}
}

// Use RouteInfo{} struct for describe router, and use ResultInfo{} to describe responses. Even you can add function to validate data  func (r *Data) AutoValidate() error
func (g *Router) Get(path string, handler interface{}, configs ...interface{}) {
	configs = append(configs, g.Info)
	g.GET(path, simpleWrapper(g.BasePath()+path, handler, "GET", checkRouter(g.BasePath(), configs...)...))
}

// Use RouteInfo{} struct for describe router, and use ResultInfo{} to describe responses. Even you can add function to validate data  func (r *Data) AutoValidate() error
func (g *Router) Post(path string, handler interface{}, configs ...interface{}) {
	configs = append(configs, g.Info)
	g.POST(path, simpleWrapper(g.BasePath()+path, handler, "POST", checkRouter(g.BasePath(), configs...)...))
}

// Use RouteInfo{} struct for describe router, and use ResultInfo{} to describe responses. Even you can add function to validate data  func (r *Data) AutoValidate() error
func (g *Router) Put(path string, handler interface{}, configs ...interface{}) {
	configs = append(configs, g.Info)
	g.PUT(path, simpleWrapper(g.BasePath()+path, handler, "PUT", checkRouter(g.BasePath(), configs...)...))
}

// Use RouteInfo{} struct for describe router, and use ResultInfo{} to describe responses. Even you can add function to validate data  func (r *Data) AutoValidate() error
func (g *Router) Delete(path string, handler interface{}, configs ...interface{}) {
	configs = append(configs, g.Info)
	g.DELETE(path, simpleWrapper(g.BasePath()+path, handler, "DELETE", checkRouter(g.BasePath(), configs...)...))
}

func checkRouter(path string, configs ...interface{}) []interface{} {

	haveTags := false
	cnf := []interface{}{}
	if path != "" && path != "/" {
		if len(strings.Split(path, "/")) > 1 {

			routerName := strings.Split(path, "/")[1]
			routerName = strings.ToUpper(routerName[:1]) + routerName[1:]

			var routeInfo RouteInfo
			for i := range configs {

				if info, ok := configs[i].(RouteInfo); ok {
					if len(info.Tags) > 0 {
						routeInfo.Tags = append(routeInfo.Tags, routerName)
						haveTags = true
					}
					if info.NeedAuthorization {
						routeInfo.NeedAuthorization = true
					}
					if info.Description != "" {
						routeInfo.Description = ""
					}
					if info.Title != "" {
						routeInfo.Title = ""
					}

				} else if info, ok := configs[i].(GroupInfo); ok {
					if len(routeInfo.Tags) == 0 {
						if info.Title != "" {
							routeInfo.Tags = append(routeInfo.Tags, info.Title)
							haveTags = true
						} else {
							routeInfo.Tags = append(routeInfo.Tags, routerName)
							haveTags = true
						}
					}

					if info.NeedAuthorization && !routeInfo.NeedAuthorization {
						routeInfo.NeedAuthorization = true
					}

				} else if info, ok := configs[i].(ResultInfo); ok {
					cnf = append(cnf, info)
				} else if info, ok := configs[i].(ResultsInfo); ok {
					for k, v := range info {
						cnf = append(cnf, ResultInfo{Code: k, Output: v})
					}
				}
			}

			if !haveTags {
				routeInfo.Tags = []string{routerName}
			}

			cnf = append(cnf, routeInfo)
		}
	}

	return cnf
}
