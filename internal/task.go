package internal

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"

	"time"
)

type task struct {
	ID     string  `json:"id"`
	Summary string  `json:"summary"`
	Date time.Time  `json:"date"`
}

func (s *server) getTasks(c *gin.Context) {
	tasks, err := s.getTasksDB()
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, "{}") //todo
		return
	}
	c.JSON(http.StatusOK, tasks)
}

func (s *server) getTasksDB() ([]task, error) {
	rows, err := s.db.Query("SELECT id, created_date FROM tasks")
	if err != nil {
		return nil, fmt.Errorf("failed to query Database: %s", err)
	}
	defer rows.Close()

	var tasks []task
	for rows.Next() {
		var task task
		err := rows.Scan(&task)
		if err != nil {
			return nil, fmt.Errorf("failed to parse task: %s", err)
		}
		tasks = append(tasks, task)
	}
	err = rows.Err()
	if err != nil {
		return nil, fmt.Errorf("failed to parse tasks: %s", err)
	}
	return tasks, nil
}


func (s *server) postTask(c *gin.Context) {
	var task task

	// Call BindJSON to bind the received JSON to
	// task.
	if err := c.BindJSON(&task); err != nil {
		return
	}

	// Add the new album to the slice.
	//albums = append(albumss, task)
	c.IndentedJSON(http.StatusCreated, task)
}