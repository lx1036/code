package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"html/template"
	"net/http"
	"time"
)

func main()  {
	router := gin.Default()
	router.Delims("{{", "}}")
	router.SetFuncMap(template.FuncMap{
		"formatAsDate": formatAsDate,
	})
	router.LoadHTMLFiles("./data/raw.tpl")
	router.GET("/raw", func(context *gin.Context) {
		context.HTML(http.StatusOK, "raw.tpl", map[string]interface{}{
			"now": time.Date(2019, time.August, 10, 0, 0, 0, 0, time.UTC),
		})
	})

	router.Run(":8080")
}

func formatAsDate(t time.Time) string  {
	year, month, day := t.Date()
	return fmt.Sprintf("%d%02d/%02d", year, month, day)
}
