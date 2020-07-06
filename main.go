package main

import (
	authn "github.com/dadrus/gin-authn"
	"github.com/gin-gonic/gin"
)

var router *gin.Engine

func main() {
	router = gin.Default()
	router.Static("/assets", "./assets")

	router.LoadHTMLGlob("templates/*")
	router.Use(authn.OAuth2Aware())

	initRoutes()

	router.Run(":8082")
}
