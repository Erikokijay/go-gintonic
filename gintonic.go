package gintonic

import (
	"encoding/base64"
	"net/http"
	"os"
	"slices"
	"strings"

	"github.com/gin-gonic/gin"
)

type Router struct {
	*gin.RouterGroup
	Info GroupInfo
}

var conf *ConfigSchema

func swaggerAuthEnabled(c *ConfigSchema) bool {
	if len(c.SwaggerUsers) > 0 {
		return true
	}
	return c.SwaggerUser != "" && c.SwaggerPassword != ""
}

func swaggerAuthValid(c *ConfigSchema, user, password string) bool {
	for _, cred := range c.SwaggerUsers {
		if cred.User == user && cred.Password == password {
			return true
		}
	}
	if c.SwaggerUser != "" && c.SwaggerPassword != "" && c.SwaggerUser == user && c.SwaggerPassword == password {
		return true
	}
	return false
}

func Config(config *ConfigSchema, eng *gin.Engine) {
	if config == nil {
		conf = &ConfigSchema{
			Mode: gin.DebugMode,
		}
	} else {
		conf = config
	}

	conf.engine = eng

	if swaggerAuthEnabled(conf) {
		eng.Use(func(c *gin.Context) {
			if !strings.HasPrefix(c.Request.URL.Path, conf.SwaggerUrl) {
				c.Next()
				return
			}
			const prefix = "Basic "
			auth := c.GetHeader("Authorization")
			if !strings.HasPrefix(auth, prefix) {
				c.Header("WWW-Authenticate", `Basic realm="Swagger"`)
				c.AbortWithStatus(http.StatusUnauthorized)
				return
			}
			decoded, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(auth, prefix))
			if err != nil {
				c.Header("WWW-Authenticate", `Basic realm="Swagger"`)
				c.AbortWithStatus(http.StatusUnauthorized)
				return
			}
			parts := strings.SplitN(string(decoded), ":", 2)
			if len(parts) != 2 || !swaggerAuthValid(conf, parts[0], parts[1]) {
				c.Header("WWW-Authenticate", `Basic realm="Swagger"`)
				c.AbortWithStatus(http.StatusUnauthorized)
				return
			}
			c.Next()
		})
	}

	if len(conf.SwaggerIPs) > 0 {
		eng.Use(func(c *gin.Context) {

			ip := c.ClientIP()

			if !slices.Contains(conf.SwaggerIPs, ip) && strings.HasPrefix(c.Request.URL.Path, conf.SwaggerUrl) {
				c.AbortWithStatus(http.StatusNotFound)
				return
			}
		})
	}

	_, err := os.Stat("./docs")
	if err != nil {
		if os.IsNotExist(err) {
			os.Mkdir("./docs", os.FileMode(0777))
		}
	}

	if conf.SwaggerUrl != "" {
		eng.StaticFile(conf.SwaggerUrl, "./docs/swagger.html")
		eng.StaticFile(conf.SwaggerUrl+"/openapi.json", "./docs/openapi.json")
	}
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

// SubGroup - create subgroup of router. It will not add routeInfo to Swagger
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
						routeInfo.Tags = append(routeInfo.Tags, info.Tags...)
						haveTags = true
					}
					if info.NeedAuthorization {
						routeInfo.NeedAuthorization = true
					}
					if info.Description != "" {
						routeInfo.Description = info.Description
					}
					if info.Title != "" {
						routeInfo.Title = info.Title
					}

				} else if info, ok := configs[i].(GroupInfo); ok {

					if info.Title != "" {
						routeInfo.Tags = append(routeInfo.Tags, info.Title)
						haveTags = true
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
