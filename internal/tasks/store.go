package tasks

func (taskService *Service) getTaskFromStore(id int) (*task, error) {
	task := &task{}
	err := taskService.DB.Get(task, "SELECT t.id, t.created_date, t.started_date, t.completed_date, u.id as 'user.id', u.username as 'user.username' FROM tasks t INNER JOIN users u on t.user_id = u.id WHERE t.id = ?;", id)
	if err != nil {
		return nil, err
	}
	return task, nil
}

func (taskService *Service) addTaskToStore(task *task) (*task, error) {
	result, err := taskService.DB.Exec("INSERT INTO tasks (user_id, created_date) VALUES (?, ?);", task.User.ID, task.CreatedDate)
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

func (taskService *Service) updateTaskInStore(task *task) (*task, error) {
	// use a pointer so it's nullable
	var userId *int
	if task.User != nil {
		userId = &task.User.ID
	} else {
		userId = nil
	}
	_, err := taskService.DB.Exec(
		// Coalesce the fields so we only update the ones that were not sent as empty to the API
		"UPDATE tasks SET user_id = COALESCE(?, user_id), created_date = COALESCE(?, created_date), started_date = COALESCE(?, started_date), completed_date = COALESCE(?, completed_date) WHERE id = ?;",
		userId, task.CreatedDate, task.StartedDate, task.CompletedDate, task.ID)
	if err != nil {
		return nil, err
	}

	return taskService.getTaskFromStore(task.ID)
}
