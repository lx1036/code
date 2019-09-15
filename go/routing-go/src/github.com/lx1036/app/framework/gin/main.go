package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
)

type Person struct {
	Name string
	Age int
}

func ginDemo()  {
	engine := gin.Default()
	iRoutes := engine.GET("/ping", func(context *gin.Context) {
		context.JSON(http.StatusOK, gin.H{
			"message": Person{
				Name: "lx1036",
				Age:  29,
			},
		})
	})

	iRoutes.GET("/hello", func(context *gin.Context) {
		context.JSON(http.StatusOK, gin.H{
			"data": "world",
		})
	})

	routes := engine.Routes()
	for _, route := range routes {
		fmt.Printf("Method: %s Path: %s Handler: %s HandlerFunc: %T \n", route.Method, route.Path, route.Handler, route.HandlerFunc)
	}

	err := engine.Run(":8080")
	if err != nil {
		fmt.Printf("uncaught error: %v", err)
	}
}

func pkgNetHttpDemo()  {
	http.HandleFunc("/hello", func(writer http.ResponseWriter, request *http.Request) {
		io.WriteString(writer, fmt.Sprintf("world from %s %s", request.Method, request.URL))
	})

	http.ListenAndServe(":9090", nil)
}

/*
https://golang.org/pkg/net/http/
 */
func main()  {
	//ginDemo()

	//pkgNetHttpDemo()
}
