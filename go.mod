module demo-app

go 1.14

require (
	github.com/dadrus/gin-authn v0.0.1
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/gin-gonic/gin v1.6.3
)

replace github.com/dadrus/gin-authn => ../gin-authn
