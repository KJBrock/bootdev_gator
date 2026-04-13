package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/KJBrock/bootdev_gator/internal/config"
)

type state struct {
	config *config.Config
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

	currentState := state{
		config: &configuration,
	}

	cmds := commands{
		supported: map[string]func(*state, command) error{},
	}

	cmds.register("login", handleLogin)

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

	err := s.config.SetUser(cmd.args[0])
	if err != nil {
		return errors.New("error setting user name")
	}

	fmt.Printf("user has been set to %s\n", cmd.args[0])
	return nil
}
