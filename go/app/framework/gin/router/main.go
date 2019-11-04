package main

import (
	"fmt"
	"github.com/julienschmidt/httprouter"
	"log"
	"net/http"
)

func BasicAuth(handle httprouter.Handle, requiredUsername string, requiredPassword string) httprouter.Handle  {
	return func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
		username, password, ok :=request.BasicAuth()

		if ok && username == requiredUsername && password == requiredPassword {
			handle(writer, request, params)
		} else {
			writer.Header().Set("WWW-Authenticate", "Basic realm=Restricted")
			http.Error(writer, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		}
	}
}


func main()  {
	router := httprouter.New()
	router2 := httprouter.New()
	test := router == router2
	fmt.Printf("router == router2: %t", test)
	router.GET("/ping", func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
		_, _ = fmt.Fprint(writer, "Pong")
	})

	router.GET("/hello/:name", func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
		_, _ = fmt.Fprintf(writer, "hello %s", params.ByName("name"))
	})

	router.GET("/src/*path", func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
		_, _ = fmt.Fprintf(writer, "file path: %s", params.ByName("path"))
	})

	username := "username"
	password := "password"

	router.GET("/protected", BasicAuth(func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
		_, _ = fmt.Fprint(writer, "Protected resource")
	}, username, password))

	log.Fatal(http.ListenAndServe(":9090", router))
}
