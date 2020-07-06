package main

import (
	"gin-app/handlers"
	"github.com/gin-gonic/gin"
)

func initRoutes() {
	router.GET("/", /*authn.RolesAllowed("openid"),*/ handlers.ShowIndexPage)
	router.GET("/article/view/:article_id", handlers.GetArticle)
	router.GET("/login", func(c *gin.Context) {
		c.Redirect(302, "/")
	})
}
