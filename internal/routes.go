package internal

import "sword-challenge/internal/users"

func (s *server) routes() {
	//TODO versioning
	s.router.GET("/health", nil) //todo

	apiV1 := s.router.Group("/v1")
	apiV1.GET("/tasks", s.getTasks)
	apiV1.POST("/tasks",s. postTask)

	users.Routes(apiV1, s.db)
}
