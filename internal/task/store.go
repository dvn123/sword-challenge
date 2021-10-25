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

func (s *Service) getTaskFromStore(id int) (*Task, error) {
	task := &Task{}
	err := s.db.Get(task, "SELECT t.id, t.completed_date, u.id as 'user.id', u.username as 'user.username' FROM tasks t INNER JOIN users u on t.user_id = u.id WHERE t.id = ?;", id)
	if err != nil {
		return nil, err
	}
	return task, nil
}

func (s *Service) addTaskToStore(task *Task) (*Task, error) {
	result, err := s.db.Exec("INSERT INTO tasks (user_id) VALUES (?);", task.User.ID)
	if err != nil {
		return nil, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}
	task.ID = int(id)

	return task, nil
}

func (s *Service) updateTaskInStore(task *Task) (*Task, error) {
	// use a pointer so it's nullable
	var userId *int
	if task.User != nil {
		userId = &task.User.ID
	} else {
		userId = nil
	}
	_, err := s.db.Exec(
		// Coalesce the fields so we only update the ones that were not sent as empty to the API
		// TODO: only set completed_date if they were explicitly set
		"UPDATE tasks SET user_id = COALESCE(?, user_id), completed_date = ? WHERE id = ?;",
		userId, task.CompletedDate, task.ID)
	if err != nil {
		return nil, err
	}

	return s.getTaskFromStore(task.ID)
}
