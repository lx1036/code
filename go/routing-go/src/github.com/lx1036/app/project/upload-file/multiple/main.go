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
		//file, err := context.FormFile("file")
		form, err := context.MultipartForm()

		if err != nil {
			context.String(http.StatusBadRequest, fmt.Sprintf("get form err: %s", err.Error()))
		}

		files := form.File["files"]

		for _, file := range files {
			filename := filepath.Base(file.Filename)

			if err := context.SaveUploadedFile(file, "assets/" + filename); err != nil {
				context.String(http.StatusBadRequest, fmt.Sprintf("can't save uploaded file: %s", err.Error()))
				return
			}
		}

		context.String(http.StatusCreated,
			fmt.Sprintf("%d Files uploaded successfully with name=%s, email=%s", len(files), name, email))
	})

	router.Run(":8080")
}
