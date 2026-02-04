package httpapi

import (
	"share-sniffer/internal/httpapi/httpconfig"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

type Server struct {
	cfg    *httpconfig.Config
	router *gin.Engine
	logger *zap.Logger
}

func NewServer(cfg *httpconfig.Config) *Server {
	// Initialize logger
	writeSyncer := zapcore.AddSync(&lumberjack.Logger{
		Filename:   "logs/httpapi.log",
		MaxSize:    10, // megabytes
		MaxBackups: 3,
		MaxAge:     28, // days
	})
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	core := zapcore.NewCore(zapcore.NewJSONEncoder(encoderConfig), writeSyncer, zap.InfoLevel)
	logger := zap.New(core)

	s := &Server{
		cfg:    cfg,
		logger: logger,
	}

	r := gin.Default() // Use default middleware (Logger, Recovery)

	r.GET("/ping", s.pingHandler)
	r.HEAD("/ping", s.pingHandler)
	r.GET("/time", s.timeHandler)
	r.HEAD("/time", s.timeHandler)
	
	// API endpoints
	r.POST("/api/check", s.checkHandler)
	r.GET("/api/version", s.versionHandler)
	r.GET("/api/home", s.homeHandler)
	r.GET("/api/support", s.supportHandler)
	r.GET("/api/help", s.helpHandler)

	s.router = r
	return s
}

func (s *Server) Run() error {
	port := s.cfg.Server.Port
	if len(port) > 0 && port[0] == ':' {
		port = port[1:]
	}
	s.logger.Info("Starting server", zap.String("port", port))
	return s.router.Run("0.0.0.0:" + port)
}
