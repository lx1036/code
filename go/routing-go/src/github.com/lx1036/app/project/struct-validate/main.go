package main

import (
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"gopkg.in/go-playground/validator.v8"
	"net/http"
	"reflect"
)

func main()  {
	route := gin.Default()

	if v, ok := binding.Validator.Engine().(* validator.Validate); ok {
		v.RegisterStructValidation(UserStructLevelValidation, User{})
	}

	route.POST("/user", validateUser)
	route.Run(":8081")
}

type User struct {
	FirstName string `json:"first_name"`
	LastName string `json:"last_name"`
	Email string `binding:"required,email"`
}

func validateUser(context *gin.Context)  {
	var user User

	if err := context.ShouldBindJSON(&user); err == nil {
		context.JSON(http.StatusOK, gin.H{
			"message": "User validated successfully",
		})
	} else {
		context.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
			"message": "User validated failed",
		})
	}
}

func UserStructLevelValidation(validator *validator.Validate, structLevel *validator.StructLevel)  {
	user := structLevel.CurrentStruct.Interface().(User)

	if len(user.FirstName) == 0 && len(user.LastName) == 0 {
		structLevel.ReportError(reflect.ValueOf(user.FirstName), "FirstName", "first_name", "fname")
		structLevel.ReportError(reflect.ValueOf(user.LastName), "LastName", "last_name", "lname")
	}
}
