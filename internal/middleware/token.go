package middleware

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"schrodinger-box/internal/misc"
	"schrodinger-box/internal/model"
)

// This is a middleware checking whether token is present and valid
func TokenMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		token := model.Token{}
		if err := ctx.ShouldBindHeader(&token); err != nil {
			misc.ReturnStandardError(ctx, http.StatusUnauthorized, "token missing")
			return
		}
		db := ctx.MustGet("DB").(*gorm.DB)
		if token.ID <= 0 {
			misc.ReturnStandardError(ctx, http.StatusBadRequest, "invalid token ID")
			return
		} else if err := db.Where(&token).First(&token).Error; errors.Is(err, gorm.ErrRecordNotFound) {
			misc.ReturnStandardError(ctx, http.StatusUnauthorized, "invalid token information")
			return
		} else if err != nil {
			misc.ReturnStandardError(ctx, http.StatusInternalServerError, err.Error())
			return
		}
		ctx.Set("Token", &token)
		// Get related user object as well
		if *token.Status == "active" {
			user := model.User{}
			user.NUSID = token.NUSID
			if err := db.Where(&user).First(&user).Error; err == nil {
				ctx.Set("User", &user)
			} else if !errors.Is(err, gorm.ErrRecordNotFound) {
				// There is something wrong other than RecordNotFound (RNF means user has not been created)
				misc.ReturnStandardError(ctx, http.StatusInternalServerError, err.Error())
				return
			}
		} else {
			misc.ReturnStandardError(ctx, http.StatusUnauthorized, "token is not active")
			return
		}
	}
}
