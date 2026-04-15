package main

import (
	"context"
	"errors"

	"github.com/KJBrock/bootdev_gator/internal/database"
)

func middlewareLoggedIn(handler func(s *state, cmd command, user database.User) error) func(*state, command) error {
	standardHandler := func(s *state, cmd command) error {
		u, err := s.db.GetUser(context.Background(), s.cfg.CurrentUserName)
		if err != nil {
			return errors.New("error finding current user")
		}

		return handler(s, cmd, u)
	}

	return standardHandler
}
