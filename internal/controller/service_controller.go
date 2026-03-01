package controller

import (
	"ne_noy/internal/repository"
	"runtime/metrics"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "ne_noy/docs"
)

const (
	nGo  = "/sched/goroutines:goroutines"
	nMem = "/memory/classes/heap/alloc:bytes"
)

type serviceController struct {
	userRepository repository.UserRepository
}

var (
	nGoroutines = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "packt",
		Name:      "goroutines",
		Help:      "Number of goroutines",
	})
	nMemory = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "packt",
		Name:      "heap_alloc_bytes",
		Help:      "Heap allocated memory",
	})
)

func init() {
	prometheus.MustRegister(nGoroutines, nMemory)
	samples := make([]metrics.Sample, 2)
	samples[0].Name = nGo
	samples[1].Name = nMem

	go func() {
		for {
			metrics.Read(samples)

			if samples[0].Value.Kind() == metrics.KindUint64 {
				nGoroutines.Set(float64(samples[0].Value.Uint64()))
			}

			switch samples[1].Value.Kind() {
			case metrics.KindUint64:
				nMemory.Set(float64(samples[1].Value.Uint64()))
			case metrics.KindFloat64:
				nMemory.Set(samples[1].Value.Float64())
			}

			time.Sleep(5 * time.Second)
		}
	}()
}

// swagger:meta
//
//	@title			Ne-Noy API
//	@version		1.0
//	@description	Backend API проекта Ne-Noy
//	@BasePath		/api
//
// AuthToken:
//
//	type: apiKey
//	name: Authorization
//	in: header
func _() {}

func (sc *serviceController) healthCheck(c *gin.Context) {
	c.JSON(200, gin.H{"status": "ok"})
}

// validateSign godoc
//
//	@Summary	Проверка подписи
//	@Tags		service
//	@Param		X-Request-Id	header	string						true	"X-Request-Id"
//	@Success	200
//	@Failure	401	{object}	dto.ErrorResponse
//	@Router		/v1/validate [get]
//	@Security	VkAuth
func validateSign(c *gin.Context) {
	c.Status(200)
}

func ConfigureServiceController(router *gin.RouterGroup, userRepository repository.UserRepository) {
	sc := &serviceController{userRepository: userRepository}
	router.GET("/health", sc.healthCheck)
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
}

func ApiServiceController(router *gin.RouterGroup) {
	router.GET("/validate", validateSign)
}
