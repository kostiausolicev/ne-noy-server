package internal

import (
	"fmt"
	vkClient "ne_noy/internal/client"
	"ne_noy/internal/config"
	"ne_noy/internal/controller"
	"ne_noy/internal/controller/middleware"
	"ne_noy/internal/repository/pgx"
	"ne_noy/internal/service"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Server struct {
	Router *gin.Engine
}

func New(db *pgxpool.Pool, config config.Config) *Server {
	userRepo := pgx.NewUserRepository(db)
	eventRepo := pgx.NewEventRepositoryPgx(db)
	roleRepo := pgx.NewRoleRepositoryPgx(db)
	eventParticipantRepository := pgx.NewEventParticipantRepository(db)
	eventQueueRepository := pgx.NewEventQueueRepository(db)

	vkCl := vkClient.NewVkApiClient(config.VK.ServiceKey, config.VK.BaseURL)

	userService := service.NewUserService(userRepo, roleRepo, vkCl)
	eventService := service.NewEventService(eventRepo, userService, roleRepo)
	eventParticipantService := service.NewEventParticipantService(eventParticipantRepository, eventRepo, config.Distance)
	// сервис для обработки callback'ов VK (добавление в очередь и т.п.)
	vkCallbackService := service.NewVkCallbackService(eventQueueRepository, eventRepo, eventParticipantService)
	// сервис для получения записей очереди
	eventQueueService := service.NewEventQueueService(eventQueueRepository)

	onboardingRepo := pgx.NewOnboardingRepository(db)
	onboardingService := service.NewOnboardingService(onboardingRepo)
	healthImportService := service.NewHealthImportService()

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
			controller.ConfigureHealthImportController(apiV1, healthImportService)
			controller.ConfigureUserController(apiV1, userService)
			apiV1.Use(middleware.AdminMiddleware())
			{
				controller.ConfigureEventQueueController(apiV1, eventQueueService)
				controller.ConfigureAdminUserController(apiV1, userService)
			}
		}
	}
	return &Server{Router: router}
}

func (s *Server) Run(host string, port int) error {
	addr := fmt.Sprintf("%s:%d", host, port)
	return s.Router.Run(addr)
}
