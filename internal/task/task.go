package task

import (
	"database/sql"
	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"net/http"
	"sword-challenge/internal/user"
	"sword-challenge/internal/util"
	"time"
)

type task struct {
	ID            int        `json:"id,omitempty"`
	Summary       string     `json:"summary,omitempty" binding:"required"`
	CompletedDate *time.Time `json:"completedDate" db:"completed_date"`
	User          *user.User `json:"user,omitempty" binding:"required"`
}

type Notification struct {
	ID            int        `json:"id" binding:"required"`
	Manager       string     `json:"manager" binding:"required"`
	CompletedDate *time.Time `json:"completedDate" binding:"required"`
	User          *user.User `json:"user" binding:"required"`
}

func (s *Service) getTasks(c *gin.Context) {
	id, err := s.mustGetTaskID(c)
	if err != nil {
		return
	}

	task, err := s.getTaskFromStore(id)
	if err == sql.ErrNoRows {
		s.logger.Infow("Failed to find task", "taskId", id)
		c.Status(http.StatusNotFound)
		return
	} else if err != nil {
		s.logger.Warnw("Failed to get task from storage", "error", err)
		c.Status(http.StatusInternalServerError) //todo error object
		return
	}

	currentUser, err := user.CheckIdsMatchIfPresentOrIsManager(c, &task.User.ID)
	if err != nil {
		c.Status(http.StatusForbidden)
		return
	}

	decryptedTask, err := s.taskEncryptor.decryptTask(task, currentUser.ID)
	if err != nil {
		s.logger.Warnw("Failed to decrypt task")
		c.Status(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, decryptedTask)
}

func (s *Service) createTask(c *gin.Context) {
	receivedTask := &task{}
	if err := c.BindJSON(receivedTask); err != nil {
		s.logger.Infow("Failed to parse task request body", "error", err)
		return
	}

	_, err := user.CheckIdsMatchIfPresentOrIsManager(c, &receivedTask.User.ID)
	if err != nil {
		c.Status(http.StatusForbidden)
		return
	}

	et, err := s.taskEncryptor.encryptTask(receivedTask)
	// Only encrypt it summary was set
	if err != nil {
		s.logger.Warnw("Failed to encrypt task")
		c.Status(http.StatusInternalServerError)
		return
	}

	id, err := s.addTaskToStore(et)
	receivedTask.ID = id
	if err != nil {
		s.logger.Warnw("Failed to add task to storage", "error", err)
		c.JSON(http.StatusInternalServerError, nil) //todo error object
		return
	}
	c.JSON(http.StatusCreated, receivedTask)
}

func (s *Service) updateTask(c *gin.Context) {
	id, err := s.mustGetTaskID(c)
	if err != nil {
		return
	}

	receivedTask := &task{}
	if err := c.BindJSON(receivedTask); err != nil {
		s.logger.Infow("Failed to parse task from body while updating", "error", err)
		return
	}
	receivedTask.ID = id

	currentUser, err := user.CheckIdsMatchIfPresentOrIsManager(c, &receivedTask.User.ID)
	if err != nil {
		c.Status(http.StatusForbidden)
		return
	}

	taskToUpdate, err := s.getTaskFromStore(receivedTask.ID)
	if err == sql.ErrNoRows {
		s.logger.Infow("Failed to find task while updating", "taskId", id)
		c.Status(http.StatusNotFound)
		return
	} else if err != nil {
		s.logger.Infow("Failed to get task while updating", "taskId", id, "error", err)
		c.Status(http.StatusInternalServerError)
		return
	}

	// Check whether the task belongs to the user making the change or if the user is manager, we can only do this after we fetch the task from the database
	if currentUser.Role.Name != util.AdminRole && taskToUpdate.User.ID != currentUser.ID {
		c.Status(http.StatusForbidden)
		return
	}

	et := &encryptedTask{ID: receivedTask.ID, User: receivedTask.User}
	// Only encrypt if summary was set
	if receivedTask.Summary != "" {
		et2, err := s.taskEncryptor.encryptTask(receivedTask)
		if err != nil {
			s.logger.Warnw("Failed to encrypt task")
			c.Status(http.StatusInternalServerError)
			return
		}
		et = et2
	}

	updatedTask, err := s.updateTaskInStore(et)
	if err != nil {
		s.logger.Warnw("Failed to update task in storage", "error", err)
		c.JSON(http.StatusInternalServerError, nil) //todo error object
		return
	}

	if taskToUpdate.CompletedDate == nil && updatedTask.CompletedDate != nil {
		go func(t encryptedTask) {
			users, err := s.userService.GetUsersByRole(util.AdminRole)
			if err != nil {
				s.logger.Warnw("Failed to get users by role when sending notification", "error", err)
				return
			}

			for _, u := range users {
				u := u
				s.taskPublisher.PublishTask(Notification{ID: t.ID, Manager: u.Username, CompletedDate: t.CompletedDate, User: t.User})
			}

		}(*updatedTask)
	}
	c.JSON(http.StatusOK, updatedTask)
}

func (s *Service) deleteTask(c *gin.Context) {
	id, err := s.mustGetTaskID(c)
	if err != nil {
		return
	}

	authUser, _ := c.Get(util.UserContextKey)
	currentUser := authUser.(*user.User)
	if currentUser.Role.Name != util.AdminRole {
		c.Status(http.StatusForbidden)
		return
	}

	rowsAffected, err := s.deleteTaskFromStore(id)
	if err != nil {
		s.logger.Infow("Failed to delete task", "taskId", id, "error", err)
		c.Status(http.StatusInternalServerError)
		return
	} else if rowsAffected == 0 {
		s.logger.Infow("Failed to find task while deleting", "taskId", id)
		c.Status(http.StatusNotFound)
		return
	}

	c.Status(http.StatusOK)
}
