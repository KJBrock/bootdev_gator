package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/KJBrock/bootdev_gator/internal/config"
	"github.com/KJBrock/bootdev_gator/internal/database"
	"github.com/google/uuid"

	_ "github.com/lib/pq"
)

type state struct {
	cfg *config.Config
	db  *database.Queries
}

type command struct {
	command string
	args    []string
}

type commands struct {
	supported map[string]func(*state, command) error
}

func (c *commands) run(s *state, cmd command) error {
	if s == nil {
		return errors.New("invalid parameter")
	}

	if f, ok := c.supported[cmd.command]; ok {
		return f(s, cmd)
	}

	return errors.New("unsupported command")
}

func (c *commands) register(name string, f func(*state, command) error) error {
	c.supported[name] = f

	return nil
}

func main() {
	configuration, err := config.Read()
	if err != nil {
		fmt.Printf("error reading config\n")
		os.Exit(1)
	}

	db, err := sql.Open("postgres", configuration.DBUrl)
	if err != nil {
		fmt.Printf("error connecting to database\n")
		os.Exit(1)
	}

	dbQueries := database.New(db)

	currentState := state{
		cfg: &configuration,
		db:  dbQueries,
	}

	cmds := commands{
		supported: map[string]func(*state, command) error{},
	}

	cmds.register("login", handleLogin)
	cmds.register("register", register)
	cmds.register("reset", resetUsers)
	cmds.register("users", getUsers)

	args := os.Args
	if len(args) < 2 {
		fmt.Printf("not enough arguments were provided\n")
		os.Exit(1)
	}

	c := command{
		command: args[1],
		args:    args[2:],
	}

	err = cmds.run(&currentState, c)
	if err != nil {
		fmt.Printf("%v\n", err)
		os.Exit(1)
	}
}

func handleLogin(s *state, cmd command) error {
	if len(cmd.args) == 0 {
		return errors.New("a username is required")
	}

	currentName := cmd.args[0]
	_, err := s.db.GetUser(context.Background(), currentName)
	if err != nil {
		return errors.New("user does not exist")
	}

	err = s.cfg.SetUser(currentName)
	if err != nil {
		return errors.New("error setting user name")
	}

	fmt.Printf("user has been set to %s\n", cmd.args[0])
	return nil
}

func register(s *state, cmd command) error {
	if len(cmd.args) == 0 {
		return errors.New("a name is required")
	}

	currentName := cmd.args[0]
	t := time.Now()
	dbUser, err := s.db.CreateUser(context.Background(),
		database.CreateUserParams{
			ID:        uuid.New(),
			CreatedAt: t,
			UpdatedAt: t,
			Name:      currentName,
		})
	if err != nil {
		return errors.New("error creating database user")
	}

	s.cfg.SetUser(currentName)

	fmt.Printf("created db user %s\n", currentName)
	fmt.Printf("user info: %v\n", dbUser)

	return nil
}

func resetUsers(s *state, cmd command) error {
	err := s.db.ResetUsers(context.Background())
	if err != nil {
		return errors.New("error resetting users")
	}

	fmt.Printf("reset users\n")

	return nil
}

func getUsers(s *state, cmd command) error {
	users, err := s.db.GetUsers(context.Background())
	if err != nil {
		return errors.New("error getting user information")
	}

	for _, u := range users {
		userName := u.Name
		if s.cfg.CurrentUserName == userName {
			userName = userName + " (current)"
		}

		fmt.Printf("%s\n", userName)
	}

	return nil

}
