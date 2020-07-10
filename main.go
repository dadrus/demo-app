package main

import (
	"errors"
	authn "github.com/dadrus/gin-authn"
	"github.com/gin-gonic/gin"
	"os"
)

var router *gin.Engine

func main() {
	router = gin.Default()

	router.LoadHTMLGlob("web/templates/*")
	router.Use(authn.OAuth2Aware())

	initRoutes()

	var port string
	if port = os.Getenv("PORT"); len(port) == 0 {
		port = "8082"
	}

	if tlsConfig, err := getTlsConfig(); err == nil {
		router.RunTLS(":" + port, tlsConfig.TlsCertFile, tlsConfig.TlsKeyFile)
	} else {
		router.Run(":" + port)
	}
}

type tlsConfig struct {
	TlsKeyFile  string
	TlsCertFile string
}

func getTlsConfig() (*tlsConfig, error) {
	tlsKeyFile := os.Getenv("TLS_KEY")
	if len(tlsKeyFile) == 0 {
		return nil, errors.New("no TLS key configured")
	}
	if _, err := os.Stat(tlsKeyFile); err != nil {
		return nil, errors.New("configured TLS key not available")
	}

	tlsCertFile := os.Getenv("TLS_CERT")
	if len(tlsCertFile) == 0 {
		return nil, errors.New("no TLS cert configured")
	}
	if _, err := os.Stat(tlsCertFile); err != nil {
		return nil, errors.New("configured TLS cert not available")
	}

	return &tlsConfig{
		TlsKeyFile:  tlsKeyFile,
		TlsCertFile: tlsCertFile,
	}, nil
}
