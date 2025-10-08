package http

import (
	"context"
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

var ErrServerClosed = http.ErrServerClosed

type HttpService struct {
	engine *gin.Engine
	server *http.Server
}

func NewHttpService(port string, isDebug bool) *HttpService {

	switch isDebug {
	case true:
		gin.SetMode(gin.DebugMode)
	default:
		gin.SetMode(gin.ReleaseMode)
	}

	g := gin.New()
	g.Use(cors.Default())
	g.Use(gin.CustomRecovery(func(c *gin.Context, err any) {
		var errMsg string
		switch v := err.(type) {
		case string:
			errMsg = v
		case error:
			errMsg = v.Error()
		}
		resp := errMsg
		c.AbortWithStatusJSON(500, resp)
	}))

	return &HttpService{
		engine: g,
		server: &http.Server{
			Addr:    ":" + port,
			Handler: g,
		},
	}
}

func (s *HttpService) Engine() *gin.Engine {
	return s.engine
}

func (s *HttpService) Start() error {
	return s.server.ListenAndServe()
}

func (s *HttpService) Stop(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}
