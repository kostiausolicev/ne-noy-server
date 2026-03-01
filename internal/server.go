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

func New(db *gorm.DB, secret string, appId int64) *Server {
	userRepo := repository.NewUserRepository(db)
	eventRepo := repository.NewEventRepository(db)
	roleRepo := repository.NewRoleRepository(db)
	eventParticipantRepository := repository.NewEventParticipantRepository(db)
	eventQueueRepository := repository.NewEventQueueRepository(db)

	userService := service.NewUserService(userRepo, roleRepo)
	eventService := service.NewEventService(eventRepo, userService, roleRepo)
	eventParticipantService := service.NewEventParticipantService(eventParticipantRepository, eventRepo)
	// сервис для обработки callback'ов VK (добавление в очередь и т.п.)
	vkCallbackService := service.NewVkCallbackService(eventQueueRepository, eventRepo, eventParticipantService)
	// сервис для получения записей очереди
	eventQueueService := service.NewEventQueueService(eventQueueRepository)

	router := gin.Default()
	// TODO Вынести в конфиг файл
	router.Use(cors.New(cors.Config{
		AllowAllOrigins:  true,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "*"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           300,
	}))
	router.Use(middleware.ErrorHandler())
	{
		public := router.Group("/")
		{
			controller.ConfigureServiceController(public, userRepo)
		}
		apiV1 := router.Group("/api/v1")
		controller.ConfigureVkCallBackController(apiV1, secret, vkCallbackService)
		apiV1.Use(middleware.AuthMiddleware(secret, appId))
		{
			controller.ApiServiceController(apiV1)
			controller.ConfigureEventController(apiV1, eventService, eventParticipantService)
			controller.ConfigureUserController(apiV1, userService)
			apiV1.Use(middleware.AdminMiddleware())
			{
				controller.ConfigureEventQueueController(apiV1, eventQueueService)
				controller.ConfigureAdminUserController(apiV1, userService)
			}
		}
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
