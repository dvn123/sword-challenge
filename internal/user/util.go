package user

import (
	"fmt"
	"sword-challenge/internal/util"
)

var ErrForbidden = fmt.Errorf("IDs do not match")

// CheckIdsMatchIfPresentOrIsManager checks whether:
//	* The ID passed is not the zero value for int and it matches the user ID
//	* The user is manager
// This helper was created since this is a common authorization check
func CheckIdsMatchIfPresentOrIsManager(userInterface interface{}, id int) (*User, error) {
	currentUser := userInterface.(*User)
	// Check whether the IDs are the same or the user is a manager
	// But only if the ID is set, 0 is the zero value for int and we assume we'll never have an ID with this value on the database
	if id != 0 && id != currentUser.ID && currentUser.Role.Name != util.AdminRole {
		return nil, ErrForbidden
	}

	return currentUser, nil
}
