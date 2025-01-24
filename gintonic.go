package gintonic

import "github.com/gin-gonic/gin"

type GinTonic struct {
	eng  *gin.Engine
	conf *ConfigSchema
	Router
}

type Router struct {
	gin.RouterGroup
}

func Config(conf *ConfigSchema, eng *gin.Engine) {
	if conf == nil {
		conf = &ConfigSchema{
			Mode:       gin.DebugMode,
			SwaggerUrl: "/docs",
		}
	}

	eng.StaticFile(conf.SwaggerUrl, "./static/swagger.html")
	eng.StaticFile("/openapi.json", "./static/openapi.json")
}

func NewServer(conf *ConfigSchema) *GinTonic {

	if conf == nil {
		conf = &ConfigSchema{
			Mode:       gin.DebugMode,
			SwaggerUrl: "/docs",
		}
	}

	res := &GinTonic{}

	res.conf = conf
	gin.SetMode(conf.Mode)
	res.eng = gin.Default()

	res.eng.StaticFile(conf.SwaggerUrl, "./static/swagger.html")
	res.eng.StaticFile("/openapi.json", "./static/openapi.json")

	return res
}

func (gnt *GinTonic) Group(path string) *Router {

	res := &Router{}

	rt := gnt.eng.Group(path)

	res.Handlers = rt.Handlers

	return res
}

func (g *GinTonic) Use(handlers gin.HandlerFunc) {
	g.eng.Use(handlers)

}

func (g *GinTonic) GET(path string, handler interface{}, configs ...interface{}) {
	g.eng.GET(path, SimpleWrapper(path, handler, "GET", configs...))
}
func (g *GinTonic) POST(path string, handler interface{}, configs ...interface{}) {
	g.eng.POST(path, SimpleWrapper(path, handler, "POST", configs...))
}
func (g *GinTonic) PUT(path string, handler interface{}, configs ...interface{}) {
	g.eng.PUT(path, SimpleWrapper(path, handler, "PUT", configs...))
}
func (g *GinTonic) DELETE(path string, handler interface{}, configs ...interface{}) {
	g.eng.DELETE(path, SimpleWrapper(path, handler, "DELETE", configs...))
}

func (g *GinTonic) Run(host string) {
	GenerateSwagger(g.conf)
	g.eng.Run(host)
}
