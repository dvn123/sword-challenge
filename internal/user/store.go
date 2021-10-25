package user

import (
	"github.com/google/uuid"
)

func (s *Service) getUserFromStore(id int) (*User, error) {
	user := &User{}
	err := s.DB.Get(
		user,
		"SELECT user.id, user.username, role.name as 'role.name', role.id as 'role.id' FROM users user LEFT JOIN roles role on user.role_id = role.id WHERE user.id = ?;",
		id)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (s *Service) addUserToStore(user *User) (*User, error) {
	result, err := s.DB.Exec("INSERT INTO users (username, role_id) VALUES (?, ?);", user.Username, user.Role.ID)
	if err != nil {
		return nil, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}
	user.ID = int(id)

	return user, nil
}

func (s *Service) authenticateUser(id int) (string, error) {
	token, err := uuid.NewUUID()
	if err != nil {
		return "", err
	}
	_, err = s.DB.Query(
		"INSERT INTO tokens (uuid, user_id, created_date) VALUES (?, (SELECT id FROM users WHERE id = ?), CURRENT_TIME);", token, id)
	if err != nil {
		return "", err
	}
	return token.String(), nil
}

func (s *Service) GetUserFromToken(token string) (*User, error) {
	user := &User{}
	err := s.DB.Get(
		user,
		"SELECT user.id, user.username, role.name as 'role.name', role.id as 'role.id' FROM users user INNER JOIN tokens t on user.id = t.user_id LEFT JOIN roles role on user.role_id = role.id WHERE t.uuid = ?;",
		token)
	if err != nil {
		return nil, err
	}
	return user, nil
}