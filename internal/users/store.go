package users

import (
	"fmt"
	"github.com/jmoiron/sqlx"
)

func getUserFromStore(db *sqlx.DB) ([]user, error) {
	rows, err := db.Query("SELECT (u.id, u.username, r.name) FROM users u INNER JOIN roles r on u.role_id = r.id")
	if err != nil {
		return nil, fmt.Errorf("failed to query Database: %s", err)
	}
	defer rows.Close()

	var users []user
	for rows.Next() {
		var user user
		err := rows.Scan(&user)
		if err != nil {
			return nil, fmt.Errorf("failed to parse user: %s", err)
		}
		users = append(users, user)
	}
	err = rows.Err()
	if err != nil {
		return nil, fmt.Errorf("failed to parse users: %s", err)
	}
	return users, nil
}


func addUserToStore(db *sqlx.DB, user *user) (*user, error) {
	result, err := db.Exec("INSERT INTO users (username, role_id) VALUES (?, ?) ", user.Username, user.Role.ID)
	if err != nil {
		return nil, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}
	user.ID = id

	return user, nil
}
