package internal

import (
	"fmt"
	"ne_noy/internal/config"
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
	router.Use(cors.New(cors.Config{
		AllowOrigins: []string{"http://localhost:5173", "http://localhost:3000"},
		AllowMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders: []string{
			"Origin",
			"Content-Length",
			"Content-Type",
			"Accept",
			"Authorization",
			// TODO убрать как будет сделана проверка через вк токен
			config.UserVkIdContextKey,
			config.UserRoleContextKey,
		},
		ExposeHeaders: []string{
			"Content-Length",
		},
		AllowCredentials: true,
	}))
	router.Use(middleware.ErrorHandler())
	public := router.Group("/")
	{
		controller.ConfigureServiceController(public, jwtService, userRepo)
	}
	apiV1 := router.Group("/api/v1")
	apiV1.Use(middleware.AuthMiddleware(secret))
	controller.ConfigureEventController(apiV1, eventService, eventParticipantService)
	controller.ConfigureUserController(apiV1, userService)

	return &Server{
		Router: router,
		DB:     db,
	}
}

func (s *Server) Run(host string, port int) error {
	addr := fmt.Sprintf("%s:%d", host, port)
	return s.Router.Run(addr)
}
