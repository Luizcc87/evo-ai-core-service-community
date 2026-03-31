package middleware

import (
	"context"
	"evo-ai-core-service/internal/config"
	"evo-ai-core-service/internal/utils/jwtutils"
	"net/http"

	"github.com/gin-gonic/gin"
)

type JWTMiddleware interface {
	GetJWTMiddleware() gin.HandlerFunc
}

type jwtMiddleware struct {
	cfgJWT       *config.JWTConfig
	cfgEvolution *config.EvolutionConfig
}

func NewJWTMiddleware(cfg *config.Config) JWTMiddleware {
	return &jwtMiddleware{
		cfgJWT:       &cfg.JWT,
		cfgEvolution: &cfg.Evolution,
	}
}

func (j *jwtMiddleware) GetJWTMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, err := jwtutils.NewJWTUtils(j.cfgJWT).Parse(c.GetHeader("Authorization"))
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}

		ctx := context.WithValue(c.Request.Context(), "sub", claims.Subject)
		ctx = context.WithValue(ctx, "email", claims.Email)
		ctx = context.WithValue(ctx, "user_id", claims.UserID)
		ctx = context.WithValue(ctx, "name", claims.Name)
		ctx = context.WithValue(ctx, "is_admin", claims.Role == "tenant")
		c.Request = c.Request.WithContext(ctx)

		c.Next()
	}
}
