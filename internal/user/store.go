package user

import (
	"github.com/google/uuid"
)

func (s *Service) authenticateUser(id int) (string, error) {
	token, err := uuid.NewUUID()
	if err != nil {
		return "", err
	}
	_, err = s.DB.Exec(
		"INSERT INTO tokens (uuid, user_id, created_date) VALUES (?, (SELECT id FROM users WHERE id = ?), CURRENT_TIME);", token, id)
	if err != nil {
		return "", err
	}
	return token.String(), nil
}

func (s *Service) GetUserByToken(token string) (*User, error) {
	user := &User{}
	err := s.DB.Get(
		user,
		"SELECT u.id, u.username, r.name as 'role.name', r.id as 'role.id' FROM users u INNER JOIN tokens t on u.id = t.user_id LEFT JOIN roles r on u.role_id = r.id WHERE t.uuid = ?;",
		token)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (s *Service) GetUsersByRole(role string) ([]User, error) {
	var users []User
	err := s.DB.Select(
		&users,
		"SELECT u.id, u.username FROM users u INNER JOIN roles r on u.role_id = r.id WHERE r.name = ?;",
		role)
	if err != nil {
		return nil, err
	}
	return users, nil
}
