package gintonic

import "github.com/gin-gonic/gin"

type GinFastApi struct {
	eng  *gin.Engine
	conf *ConfigSchema
}

func NewServer(conf *ConfigSchema) *GinFastApi {

	if conf == nil {
		conf = &ConfigSchema{
			Mode:       gin.DebugMode,
			SwaggerUrl: "/docs",
		}
	}

	res := &GinFastApi{}

	res.conf = conf
	gin.SetMode(conf.Mode)
	res.eng = gin.Default()

	res.eng.StaticFile(conf.SwaggerUrl, "./static/swagger.html")
	res.eng.StaticFile("/openapi.json", "./static/openapi.json")

	return res
}

func (g *GinFastApi) Use(handlers gin.HandlerFunc) {
	g.eng.Use(handlers)
}

func (g *GinFastApi) GET(path string, handler interface{}, configs ...interface{}) {
	g.eng.GET(path, WrapHandler(handler, "GET", configs))
}
func (g *GinFastApi) POST(path string, handler interface{}, configs ...interface{}) {
	g.eng.POST(path, WrapHandler(handler, "POST", configs))
}
func (g *GinFastApi) PUT(path string, handler interface{}, configs ...interface{}) {
	g.eng.PUT(path, WrapHandler(handler, "PUT", configs))
}
func (g *GinFastApi) DELETE(path string, handler interface{}, configs ...interface{}) {
	g.eng.DELETE(path, WrapHandler(handler, "DELETE", configs))
}

func (g *GinFastApi) Run(host string) {
	generateSwagger(g.conf)
	g.eng.Run(host)
}
