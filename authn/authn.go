package authn

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	jose "github.com/square/go-jose"
	"net/http"
	"strings"
	"sync"
)

type Configuration struct {
	VerifyToken bool
	DiscoveryUrl string
	JwksUrl string
}

func OAuth2Aware(verifyToken bool) gin.HandlerFunc {
	cache := newCache()

	getTokenValidationKey := func(token *jwt.Token) (interface{}, error) {
		claims := token.Claims.(jwt.MapClaims)
		issuer := claims["iss"].(string)
		kid := token.Header["kid"].(string)

		cacheKey := issuer + kid
		jwk, present := cache.get(cacheKey)
		if !present {
			// retrieve signing key
			jwks, err := downloadJwks(issuer + ".well-known/jwks.json")
			if err != nil {
				return nil, err
			}

			jwksEntry := jwks.Key(kid)
			if len(jwksEntry) == 0 {
				return nil, errors.New("No key found for given key id")
			}

			if jwksEntry[0].Algorithm != token.Header["alg"].(string) {
				return nil, errors.New("Algorithm mismatch between token header and the algorithm n jwk")
			}

			// TODO: Check the use claim from the jwk. It shall at least contain "sig"

			jwk = &jwksEntry[0]
			cache.put(cacheKey, jwk)
		}

		// TODO: check whether certificates are present and check the chain for validity
		// if not present try to download the chain and check it then

		return jwk.Key, nil
	}

	return func(c *gin.Context) {
		if token, err := getAccessToken(c, verifyToken, getTokenValidationKey); err == nil {
			handleAccessToken(c, token)
		}

		if token, err := getIdToken(c, verifyToken, getTokenValidationKey); err == nil {
			handleIdToken(c, token)
		}

		c.Next()
	}
}

func handleAccessToken(c *gin.Context, token *jwt.Token) {
	c.Set("access_token", token)

	claims := token.Claims.(jwt.MapClaims)
	scopes, present := claims["scp"]
	if !present {
		scopes, present = claims["scope"]
	}

	if present {
		c.Set("roles", toStringSlice(scopes))
	}

	subject, present := claims["sub"]
	if present {
		c.Set("subject", subject)
	}
}

func handleIdToken(c *gin.Context, token *jwt.Token) {
	c.Set("id_token", token)
}

func getAccessToken(c *gin.Context, verifyToken bool, getTokenValidationKey func(token *jwt.Token) (interface{}, error)) (*jwt.Token, error) {
	rawToken, err := extractAccessToken(c)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	var token *jwt.Token
	if verifyToken {
		token, err = jwt.Parse(rawToken, getTokenValidationKey)
	} else {
		token, _, err = new(jwt.Parser).ParseUnverified(rawToken, jwt.MapClaims{})
	}
	return token, err
}

func getIdToken(c *gin.Context, verifyToken bool, getTokenValidationKey func(token *jwt.Token) (interface{}, error)) (*jwt.Token, error) {
	rawToken, err := extractIdToken(c)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	var token *jwt.Token
	if verifyToken {
		token, err = jwt.Parse(rawToken, getTokenValidationKey)
	} else {
		token, _, err = new(jwt.Parser).ParseUnverified(rawToken, jwt.MapClaims{})
	}
	return token, err
}

func downloadJwks(jwksUrl string) (*jose.JSONWebKeySet, error) {
	resp, err := http.Get(jwksUrl)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var jwks jose.JSONWebKeySet
	if err := json.NewDecoder(resp.Body).Decode(&jwks); err != nil {
		return nil, err
	}

	return &jwks, nil
}

func extractAccessToken(c *gin.Context) (string, error) {
	if authHeader := c.GetHeader("Authorization"); len(authHeader) != 0 &&
		strings.Index(strings.ToLower(authHeader), "bearer ") != -1 {
		pos := strings.Index(strings.ToLower(authHeader), "bearer ")
		adjPos := pos + len("bearer ")
		if adjPos >= len(authHeader) {
			return "", errors.New("Malformed Authorization header")
		} else {
			return authHeader[adjPos:], nil
		}
	} else if authBody := c.PostForm("access_token"); len(authBody) != 0 {
		return authBody, nil
	} else if authQuery := c.Query("access_token"); len(authQuery) != 0 {
		return authQuery, nil
	} else {
		return "", errors.New("No Authorization token present")
	}
}

func extractIdToken(c *gin.Context) (string, error) {
	if idHeader := c.GetHeader("X-Id-Token"); len(idHeader) != 0 {
		return idHeader, nil
	} else {
		return "", errors.New("No Id token present")
	}
}

func DenyAll() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.AbortWithError(http.StatusUnauthorized, errors.New("Not allowed"))
		return
	}
}

func RolesAllowed(allowedRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if len(allowedRoles) == 0 {
			return
		}

		roles, present := c.Get("roles")
		if !present {
			c.AbortWithError(http.StatusUnauthorized, errors.New("Not authorized"))
			return
		}

		if !containsAll(roles.([]string), allowedRoles) {
			c.AbortWithError(http.StatusUnauthorized, errors.New("Not authorized"))
			return
		}

		fmt.Println("Authenticated")
	}
}

func toStringSlice(ivalues interface{}) []string {
	svalues := ivalues.([]interface{})
	res := make([]string, len(svalues))
	for i, v := range svalues {
		res[i] = v.(string)
	}
	return res
}

func containsAll(src []string, vals []string) bool {
	if len(vals) == 0 {
		return true
	}

	for _, v := range vals {
		if !contains(src, v) {
			return false
		}
	}

	return true
}

func contains(src []string, val string) bool {
	for _, e := range src {
		if e == val {
			return true
		}
	}
	return false
}

type FilterFunc func(v string) bool

func Filter(src []string, f FilterFunc) []string {
	res := make([]string, 0)
	for _, v := range src {
		if f(v) {
			res = append(res, v)
		}
	}
	return res
}

type cache struct {
	entries map[string]*jose.JSONWebKey
	mutex   *sync.RWMutex
}

func newCache() *cache {
	return &cache{
		entries: make(map[string]*jose.JSONWebKey),
		mutex:   &sync.RWMutex{},
	}
}

func (c *cache) get(key string) (*jose.JSONWebKey, bool) {
	c.mutex.RLock()
	val, p := c.entries[key]
	c.mutex.RUnlock()
	return val, p
}

func (c *cache) put(key string, value *jose.JSONWebKey) {
	c.mutex.Lock()
	c.entries[key] = value
	c.mutex.Unlock()
}
