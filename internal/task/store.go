package task

func (s *Service) deleteTaskFromStore(id int) (int, error) {
	res, err := s.db.Exec("DELETE FROM tasks t WHERE t.id = ?;", id)
	if err != nil {
		return 0, err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return 0, err
	}
	return int(affected), nil
}

func (s *Service) getTaskFromStore(id int) (*encryptedTask, error) {
	task := &encryptedTask{}
	err := s.db.Get(task, "SELECT t.id, t.summary, t.completed_date, u.id as 'user.id', u.username as 'user.username' FROM tasks t INNER JOIN users u on t.user_id = u.id WHERE t.id = ?;", id)
	if err != nil {
		return nil, err
	}
	return task, nil
}

func (s *Service) getTasksFromStore(id int) ([]encryptedTask, error) {
	task := []encryptedTask{}
	err := s.db.Select(&task, "SELECT t.id, t.summary, t.completed_date, u.id as 'user.id', u.username as 'user.username' FROM tasks t INNER JOIN users u on t.user_id = u.id WHERE t.user_id = ?;", id)
	if err != nil {
		return nil, err
	}
	return task, nil
}

func (s *Service) addTaskToStore(task *encryptedTask) (int, error) {
	result, err := s.db.Exec("INSERT INTO tasks (user_id, summary) VALUES (?, ?);", task.User.ID, task.EncryptedSummary)
	if err != nil {
		return 0, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return int(id), nil
}

func (s *Service) updateTaskInStore(task *encryptedTask) (*encryptedTask, error) {
	_, err := s.db.Exec(
		// Coalesce the fields so we only update the ones that were not sent as empty to the API
		"UPDATE tasks SET user_id = COALESCE(?, user_id), summary = COALESCE(?, summary), completed_date = ? WHERE id = ?;",
		task.User.ID, task.EncryptedSummary, task.CompletedDate, task.ID)
	if err != nil {
		return nil, err
	}

	return s.getTaskFromStore(task.ID)
}
