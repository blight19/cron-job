package api

import (
	"dbsmonitor/crontab/master/api/auth"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
)

func JWTAuth() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var parts []string
		authHeader := ctx.Request.Header.Get("Authorization")
		if authHeader == "" {
			goto ret
		}
		parts = strings.SplitN(authHeader, " ", 2)
		if !(len(parts) == 2 && parts[0] == "Bearer") {
			fmt.Println("222")
			goto ret
		}
		if claims, err := auth.ParseToken(parts[1]); err != nil {
			fmt.Println("111")
			goto ret
		} else {
			if auth.CheckUserInfo(claims) {
				ctx.Set("claims", claims)
				ctx.Next()
				return
			}
		}
	ret:
		{
			ctx.JSON(http.StatusOK, gin.H{
				"code":    -1,
				"message": "invalid token",
			})
			ctx.Abort()
			return
		}
	}
}
