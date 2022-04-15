package helper

import (
	"errors"
	"log"

	"github.com/gin-gonic/gin"
)

//CheckUserType renews the user tokens when they login
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

//MatchUserTypeToUid only allows the user to access their data and no other data. Only the admin can access all user data
func MatchUserTypeAndUid(c *gin.Context, userId string) (err error) {
	userType := c.GetString("user_type")
	uid := c.GetString("uid")
	err = nil

	if userType == "normal" && uid != userId {
		err = errors.New("unauthorized to access this resource")
		log.Printf("unauthorized to access this resource")
		return err
	}
	err = CheckUserType(c, userType)

	return err
}

func MatchUserTypeAndName(c *gin.Context, userEmail string) (err error) {
	userType := c.GetString("user_type")
	email := c.GetString("email")
	err = nil

	if userType == "normal" && email != userEmail {
		err = errors.New("unauthorized to access this resource")
		log.Printf("unauthorized to access this resource")
		return err
	}
	err = CheckUserType(c, userType)

	return err
}
