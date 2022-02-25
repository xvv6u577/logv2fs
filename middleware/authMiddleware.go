package middleware

import (
	"net/http"

	helper "github.com/caster8013/logv2rayfullstack/helpers"

	"github.com/gin-gonic/gin"
)

// Authz validates token and authorizes users
func Authentication() gin.HandlerFunc {
	return func(c *gin.Context) {
		clientToken := c.Request.Header.Get("token")
		if clientToken == "" {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "No Authorization header provided"})
			c.Abort()
			return
		}

		claims, err := helper.ValidateToken(clientToken)
		if err != "" {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err})
			c.Abort()
			return
		}

		c.Set("email", claims.Email)
		c.Set("uuid", claims.UUID)
		c.Set("path", claims.Path)
		c.Set("uid", claims.Uid)
		c.Set("user_type", claims.Role)

		c.Next()

	}
}

func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		// c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		// c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		// c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		// c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT")

		// if c.Request.Method == "OPTIONS" {
		// 	c.AbortWithStatus(204)
		// 	return
		// }

		// c.Next()
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Header("Access-Control-Allow-Methods", "POST,HEAD,PATCH, OPTIONS, GET, PUT")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
