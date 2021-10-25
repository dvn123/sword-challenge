package task

import (
	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"net/http"
	"strconv"
	"sword-challenge/internal/user"
	"sword-challenge/internal/util"
	"time"
)

type task struct {
	ID            int        `json:"id,omitempty"`
	Summary       string     `json:"summary,omitempty" binding:"required"`
	CompletedDate *time.Time `json:"completedDate" db:"completed_date"`
	User          *user.User `json:"user,omitempty"`
}

type NotificationTask struct {
	ID            int        `json:"id" binding:"required"`
	Manager       string     `json:"manager" binding:"required"`
	CompletedDate *time.Time `json:"completedDate" binding:"required"`
	User          *user.User `json:"user" binding:"required"`
}

func (s *Service) getTasks(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("task-id"))
	if err != nil {
		s.logger.Infow("Failed to parse task ID", "error", err)
		c.Status(http.StatusBadRequest) //todo error
		return
	}
	task, err := s.getTaskFromStore(id)
	if err != nil {
		s.logger.Warnw("Failed to get task from storage", "error", err)
		c.Status(http.StatusInternalServerError) //todo error object
		return
	}

	authUser, _ := c.Get(util.UserContextKey)
	_, err = user.CheckIdsMatchIfPresentOrIsManager(authUser, &id)
	if err != nil {
		c.Status(http.StatusForbidden)
		return
	}

	c.JSON(http.StatusOK, task)
}

func (s *Service) createTask(c *gin.Context) {
	receivedTask := &task{}
	if err := c.BindJSON(receivedTask); err != nil {
		s.logger.Infow("Failed to parse task request body", "error", err)
		return
	}

	authUser, _ := c.Get(util.UserContextKey)

	var userId *int
	if receivedTask.User != nil {
		userId = &receivedTask.User.ID
	} else {
		userId = nil
	}
	_, err := user.CheckIdsMatchIfPresentOrIsManager(authUser, userId)
	if err != nil {
		c.Status(http.StatusForbidden)
		return
	}

	receivedTask, err = s.addTaskToStore(receivedTask)
	if err != nil {
		s.logger.Warnw("Failed to add task to storage", "error", err)
		c.JSON(http.StatusInternalServerError, nil) //todo error object
		return
	}
	c.JSON(http.StatusCreated, receivedTask)
}

func (s *Service) updateTask(c *gin.Context) {
	receivedTask := &task{}
	id, err := strconv.Atoi(c.Param("task-id"))
	if err != nil {
		s.logger.Infow("Failed to parse task ID", "error", err)
		c.JSON(http.StatusBadRequest, nil)
		return
	}

	if err := c.BindJSON(receivedTask); err != nil {
		s.logger.Infow("Failed to parse task from body while updating", "error", err)
		return
	}
	receivedTask.ID = id

	authUser, _ := c.Get(util.UserContextKey)
	// So we don't have a null pointer
	var userId *int
	if receivedTask.User != nil {
		userId = &receivedTask.User.ID
	} else {
		userId = nil
	}
	currentUser, err := user.CheckIdsMatchIfPresentOrIsManager(authUser, userId)
	if err != nil {
		c.Status(http.StatusForbidden)
		return
	}

	taskToUpdate, err := s.getTaskFromStore(receivedTask.ID)
	if err != nil {
		s.logger.Infow("Failed to get task while updating", "taskId", id, "error", err)
		c.Status(http.StatusInternalServerError)
		return
	} else if taskToUpdate == nil {
		s.logger.Infow("Failed to find task while updating", "taskId", id)
		c.Status(http.StatusNotFound)
		return
	}

	// Check whether the task belongs to the user making the change or if the user is manager, we can only do this after we fetch the task from the database
	// We could also do this by checking
	if currentUser.Role.Name != util.AdminRole && taskToUpdate.User.ID != currentUser.ID {
		c.Status(http.StatusForbidden)
		return
	}

	updatedTask, err := s.updateTaskInStore(receivedTask)
	if err != nil {
		s.logger.Warnw("Failed to update task in storage", "error", err)
		c.JSON(http.StatusInternalServerError, nil) //todo error object
		return
	}

	if taskToUpdate.CompletedDate == nil && updatedTask.CompletedDate != nil {
		go func(t task) {
			users, err := s.userService.GetUsersByRole(util.AdminRole)
			if err != nil {
				s.logger.Warnw("Failed to get users by role when sending notification", "error", err)
				return
			}

			for _, u := range users {
				u := u
				s.taskPublisher.PublishTask(NotificationTask{ID: t.ID, Manager: u.Username, CompletedDate: t.CompletedDate, User: t.User})
			}

		}(*updatedTask)
	}
	c.JSON(http.StatusOK, updatedTask)
}

func (s *Service) deleteTask(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("task-id"))
	if err != nil {
		s.logger.Infow("Failed to parse task ID", "error", err)
		c.JSON(http.StatusBadRequest, nil)
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
