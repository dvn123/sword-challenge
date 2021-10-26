package user

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"sword-challenge/internal/util"
)

var ErrForbidden = fmt.Errorf("IDs do not match")

// CheckIdsMatchIfPresentOrIsManager checks whether:
//	* The ID passed is not nil and it matches the user ID
//	* The user is manager
// This helper was created since this is a common authorization check
func CheckIdsMatchIfPresentOrIsManager(c *gin.Context, id *int) (*User, error) {
	uInterface, _ := c.Get(util.UserContextKey)
	currentUser := uInterface.(*User)

	if id != nil && *id != currentUser.ID && currentUser.Role.Name != util.AdminRole {
		return nil, ErrForbidden
	}

	return currentUser, nil
}
