package controller

import "github.com/gin-gonic/gin"

type serviceController struct {
}

func (sc *serviceController) generateToken(c *gin.Context) {

}

func (sc *serviceController) healthCheck(c *gin.Context) {
	c.JSON(200, gin.H{"status": "ok"})
}

func ConfigureServiceController(
	router *gin.RouterGroup) {
	sc := &serviceController{}
	router.GET("/health", func(c *gin.Context) { sc.healthCheck(c) })
}
