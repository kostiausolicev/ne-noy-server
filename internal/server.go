package internal

import (
	"fmt"
	vkClient "ne_noy/internal/client"
	"ne_noy/internal/config"
	"ne_noy/internal/controller"
	"ne_noy/internal/controller/event"
	"ne_noy/internal/controller/middleware"
	"ne_noy/internal/repository/impl"
	"ne_noy/internal/service"
	event2 "ne_noy/internal/service/event"
	"ne_noy/internal/service/event/event_as_event"
	"ne_noy/internal/service/event/event_as_team"
	"ne_noy/internal/service/event/event_as_test"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Server struct {
	Router *gin.Engine
}

func New(db *pgxpool.Pool, config config.Config) *Server {
	userRepo := impl.NewUserRepository(db)
	eventRepo := impl.NewEventBaseRepository(db)
	eventAsEventRepo := impl.NewEventEventRepository(db)
	eventTeamRepo := impl.NewEventTeamRepository(db)
	eventTestRepo := impl.NewEventTestRepository(db)
	roleRepo := impl.NewRoleRepositoryPgx(db)
	eventParticipantRepository := impl.NewEventParticipantRepository(db)
	eventQueueRepository := impl.NewEventQueueRepository(db)

	vkCl := vkClient.NewVkApiClient(config.VK.ServiceKey, config.VK.BaseURL)

	userService := service.NewUserService(userRepo, roleRepo, vkCl)
	eventService := event2.NewEventService(eventRepo, userService, roleRepo)
	eventParticipantService := event_as_event.NewEventParticipantService(eventParticipantRepository, eventAsEventRepo, config.Distance)
	eventTeamService := event_as_team.NewEventTeamService(eventTeamRepo, vkCl)
	eventTestService := event_as_test.NewEventTestService(eventTestRepo)
	// сервис для обработки callback'ов VK (добавление в очередь и т.п.)
	vkCallbackService := service.NewVkCallbackService(eventQueueRepository, eventAsEventRepo, eventParticipantService)
	// сервис для получения записей очереди
	eventQueueService := service.NewEventQueueService(eventQueueRepository)

	onboardingRepo := impl.NewOnboardingRepository(db)
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
			event.ConfigureEventController(apiV1, eventService, eventParticipantService)
			event.ConfigureTeamEventController(apiV1, eventService, eventTeamService)
			event.ConfigureTestController(apiV1, eventService, eventTestService)
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
