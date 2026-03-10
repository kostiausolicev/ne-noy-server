package internal

import (
	"fmt"
	vkClient "ne_noy/internal/client"
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

func New(db *gorm.DB, config config.Config) *Server {
	userRepo := repository.NewUserRepository(db)
	eventRepo := repository.NewEventRepository(db)
	roleRepo := repository.NewRoleRepository(db)
	eventParticipantRepository := repository.NewEventParticipantRepository(db)
	eventQueueRepository := repository.NewEventQueueRepository(db)

	vkCl := vkClient.NewVkApiClient(config.VK.ServiceKey, config.VK.BaseURL)

	userService := service.NewUserService(userRepo, roleRepo, vkCl)
	eventService := service.NewEventService(eventRepo, userService, roleRepo)
	eventParticipantService := service.NewEventParticipantService(eventParticipantRepository, eventRepo)
	// сервис для обработки callback'ов VK (добавление в очередь и т.п.)
	vkCallbackService := service.NewVkCallbackService(eventQueueRepository, eventRepo, eventParticipantService)
	// сервис для получения записей очереди
	eventQueueService := service.NewEventQueueService(eventQueueRepository)

	onboardingRepo := repository.NewOnboardingRepository(db)
	onboardingService := service.NewOnboardingService(onboardingRepo)

	sRouter := gin.New()
	public := sRouter.Group("/")
	{
		controller.ConfigureServiceController(public, userRepo)
	}

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
		apiV1 := router.Group("/api/v1")
		controller.ConfigureVkCallBackController(apiV1, config.Secret, vkCallbackService)
		apiV1.Use(middleware.AuthMiddleware(config.Secret, config.AppId))
		{
			controller.ConfigureOnboardingController(apiV1, onboardingService)
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
