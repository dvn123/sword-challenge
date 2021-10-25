package user

import (
	"fmt"
	"sword-challenge/internal/util"
)

var ErrForbidden = fmt.Errorf("IDs do not match")

// CheckIdsMatchIfPresentOrIsManager checks whether:
//	* The ID passed is not nil and it matches the user ID
//	* The user is manager
// This helper was created since this is a common authorization check
func CheckIdsMatchIfPresentOrIsManager(userInterface interface{}, id *int) (*User, error) {
	currentUser := userInterface.(*User)

	if id != nil && *id != currentUser.ID && currentUser.Role.Name != util.AdminRole {
		return nil, ErrForbidden
	}

	return currentUser, nil
}
