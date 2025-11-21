package internal

import (
	"fmt"
	"ne_noy/internal/controller"
	"ne_noy/internal/controller/middleware"
	"ne_noy/internal/repository"
	"ne_noy/internal/service"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Server struct {
	Router *gin.Engine
	DB     *gorm.DB
}

func New(db *gorm.DB, secret string) *Server {
	router := gin.Default()
	router.Use(middleware.ErrorHandler())
	router.Use(middleware.AuthMiddleware(secret))

	// healthcheck
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	userRepo := repository.NewUserRepository(db)
	eventRepo := repository.NewEventRepository(db)
	eventParticipantRepository := repository.NewEventParticipantRepository(db)

	userService := service.NewUserService(userRepo)
	eventService := service.NewEventService(eventRepo)
	eventParticipantService := service.NewEventParticipantService(eventParticipantRepository)

	controller.ConfigureUserController(router, userService)
	controller.ConfigureEventController(router, eventService, eventParticipantService)

	return &Server{
		Router: router,
		DB:     db,
	}
}

func (s *Server) Run(host string, port int) error {
	addr := fmt.Sprintf("%s:%d", host, port)
	return s.Router.Run(addr)
}
