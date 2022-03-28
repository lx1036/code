package http

import "github.com/gin-gonic/gin"

func main() {
	router := gin.Default()
	router.GET("ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "It's OK",
		})
	})

	//router.POST("somepost", posting)

	router.Run(":8099")
}
