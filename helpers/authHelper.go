package helper

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

// CheckUserType renews the user tokens when they login
func CheckUserType(c *gin.Context, role string) (err error) {
	userType := c.GetString("user_type")
	err = nil
	if userType != role {
		err = errors.New("unauthorized to access this resource")
		log.Printf("unauthorized to access this resource")
		return err
	}

	return err
}

func MatchUserTypeAndName(c *gin.Context, userEmail string) (err error) {
	userType := c.GetString("user_type")
	email := c.GetString("email")
	err = nil

	if userType == "normal" && email != userEmail {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if !SameUserID(ctx, email, userEmail) {
			err = errors.New("unauthorized to access this resource")
			log.Printf("unauthorized to access this resource")
			return err
		}
	}
	err = CheckUserType(c, userType)

	return err
}
