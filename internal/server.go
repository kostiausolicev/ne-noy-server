package internal

import (
	"fmt"
	"ne_noy/internal/controller"
	"ne_noy/internal/controller/middleware"
	"ne_noy/internal/repository"
	"ne_noy/internal/service"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Server struct {
	Router *gin.Engine
	DB     *gorm.DB
}

func New(db *gorm.DB, secret string) *Server {
	userRepo := repository.NewUserRepository(db)
	eventRepo := repository.NewEventRepository(db)
	eventParticipantRepository := repository.NewEventParticipantRepository(db)

	userService := service.NewUserService(userRepo)
	eventService := service.NewEventService(eventRepo)
	eventParticipantService := service.NewEventParticipantService(eventParticipantRepository, eventRepo)
	jwtService := service.NewJWTService(secret)

	router := gin.Default()
	router.Use(cors.Default())
	router.Use(middleware.ErrorHandler())
	public := router.Group("/")
	{
		controller.ConfigureServiceController(public, jwtService, userRepo)
		controller.ConfigureUserController(public, userService)
	}
	apiV1 := router.Group("/api/v1")
	apiV1.Use(middleware.AuthMiddleware(secret))
	{
		controller.ConfigureEventController(apiV1, eventService, eventParticipantService)
	}

	return &Server{
		Router: router,
		DB:     db,
	}
}

func (s *Server) Run(host string, port int) error {
	addr := fmt.Sprintf("%s:%d", host, port)
	return s.Router.Run(addr)
}
