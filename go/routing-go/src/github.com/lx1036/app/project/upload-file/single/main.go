package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"path/filepath"
)

func main()  {
	router := gin.Default()
	router.MaxMultipartMemory = 8 << 20 // 8MB
	router.Static("/", "./public")
	router.POST("/upload", func(context *gin.Context) {
		name := context.PostForm("name")
		email := context.PostForm("email")
		file, err := context.FormFile("file")

		if err != nil {
			context.String(http.StatusBadRequest, fmt.Sprintf("form file error: %s", err.Error()))
			return
		}

		filename := filepath.Base(file.Filename)
		if err := context.SaveUploadedFile(file, "assets/" + filename); err != nil {
			context.String(http.StatusBadRequest, fmt.Sprintf("can't save uploaded file: %s", err.Error()))
			return
		}

		context.String(http.StatusCreated,
			fmt.Sprintf("File %s uploaded successfully with name=%s, email=%s", file.Filename, name, email))
	})

	router.Run(":8080")
}
