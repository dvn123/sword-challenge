package user

import (
	"fmt"
	"sword-challenge/internal/util"
)

var ErrForbidden = fmt.Errorf("IDs do not match")

func CheckIdsMatchOrIsManager(userInterface interface{}, id int) (*User, error) {
	currentUser := userInterface.(*User)
	// Check whether the IDs are the same or the user is a manager
	// But only if the ID is set, 0 is the zero value for int and we assume we'll never have an ID with this value on the database
	if id != 0 && id != currentUser.ID && currentUser.Role.Name != util.AdminRole {
		return nil, ErrForbidden
	}

	return currentUser, nil
}

// Helper for null checking
func GetIDOrZeroValue(u *User) int {
	if u == nil {
		return 0
	} else {
		return u.ID
	}
}
