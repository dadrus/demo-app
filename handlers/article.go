package handlers

import (
	"demo-app/models"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

// Render one of HTML, JSON or CSV based on the 'Accept' header of the request
// If the header doesn't specify this, HTML is rendered, provided that
// the template name is present
func render(c *gin.Context, data gin.H, templateName string) {
	switch c.Request.Header.Get("Accept") {
	case "application/json":
		// Respond with JSON
		c.JSON(http.StatusOK, data["payload"])
	case "application/xml":
		// Respond with XML
		c.XML(http.StatusOK, data["payload"])
	default:
		// Respond with HTML
		c.HTML(http.StatusOK, templateName, data)
	}
}

func ShowIndexPage(c *gin.Context) {

	var userClaims map[string]interface{}
	if token, present := c.Get("id_token"); present {
		userClaims = token.(*jwt.Token).Claims.(jwt.MapClaims)
	}

	articles := models.GetAllArticles()

	var obj gin.H
	if len(userClaims) != 0 {
		obj = gin.H{"title": "Home Page 2", "user": userClaims, "payload": articles}
	} else {
		obj = gin.H{"title": "Home Page 2", "payload": articles}
	}

	// Call the render function with the name of the template to render
	c.HTML(http.StatusOK, "index.html", obj)
}

func GetArticle(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("article_id"))
	if err != nil {
		c.AbortWithError(http.StatusNotFound, err)
		return
	}

	article, err := models.GetArticleById(id)
	if err != nil {
		c.AbortWithError(http.StatusNotFound, err)
		return
	}

	var userClaims map[string]interface{}
	if token, present := c.Get("id_token"); present {
		userClaims = token.(*jwt.Token).Claims.(jwt.MapClaims)
	}

	// Call the render function with the name of the template to render
	render(c,
		gin.H{
			"title":   article.Title,
			"user":    userClaims,
			"payload": article},
		"article.html")

}
