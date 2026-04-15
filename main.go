package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/KJBrock/bootdev_gator/internal/config"
	"github.com/KJBrock/bootdev_gator/internal/database"
	"github.com/google/uuid"

	_ "github.com/lib/pq"
)

const testFeed = "https://www.wagslane.dev/index.xml"

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
	cmds.register("agg", aggregateFeeds)
	cmds.register("feeds", getFeeds)
	cmds.register("addfeed", middlewareLoggedIn(addFeed))
	cmds.register("follow", middlewareLoggedIn(followFeed))
	cmds.register("following", middlewareLoggedIn(followingFeeds))
	cmds.register("unfollow", middlewareLoggedIn(unfollowFeed))
	cmds.register("browse", middlewareLoggedIn(browsePosts))

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
		fmt.Printf("Error: %v\n", err)
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

func aggregateFeeds(s *state, cmd command) error {
	if len(cmd.args) == 0 {
		return errors.New("please specify the query interval")
	}

	fmt.Printf("Collecting feeds every %s\n", cmd.args[0])

	timeBetweenRequests, err := time.ParseDuration(cmd.args[0])
	if err != nil {
		return err
	}

	ticker := time.NewTicker(timeBetweenRequests)
	for ; ; <-ticker.C {
		scrapeError := scrapeFeeds(s)
		if scrapeError != nil {
			fmt.Printf("error scraping feeds: %v\n", scrapeError)
			break
		}
	}

	fmt.Printf("exiting aggregation loop\n")

	return nil
}

func getFeeds(s *state, _ command) error {
	feeds, err := s.db.GetFeeds(context.Background())
	if err != nil {
		return errors.New("error getting user information")
	}

	for _, feed := range feeds {
		user, userErr := s.db.GetUserByID(context.Background(), feed.UserID)
		if userErr != nil {
			return userErr
		}

		fmt.Printf("Name: %s, URL: %s, Created By: %s\n", feed.Name, feed.Url, user.Name)
	}

	return nil

}

func addFeed(s *state, cmd command, user database.User) error {
	if len(cmd.args) < 2 {
		return errors.New("a name and URL are required")
	}

	feedName := cmd.args[0]
	feedURL := cmd.args[1]

	t := time.Now()
	feed, err := s.db.CreateFeed(context.Background(),
		database.CreateFeedParams{
			ID:        uuid.New(),
			CreatedAt: t,
			UpdatedAt: t,
			Name:      feedName,
			Url:       feedURL,
			UserID:    user.ID,
		})
	if err != nil {
		return err
	}

	_, err = s.db.CreateFeedFollow(context.Background(),
		database.CreateFeedFollowParams{
			ID:        uuid.New(),
			CreatedAt: t,
			UpdatedAt: t,
			FeedID:    feed.ID,
			UserID:    user.ID,
		})

	fmt.Printf("%v\n", feed)

	return nil
}

func followFeed(s *state, cmd command, user database.User) error {
	if len(cmd.args) != 1 {
		return errors.New("specify a feed URL")
	}

	f, err := s.db.GetFeedByUrl(context.Background(), cmd.args[0])
	if err != nil {
		return errors.New("error finding feed by URL")
	}

	t := time.Now()
	ff, err := s.db.CreateFeedFollow(context.Background(),
		database.CreateFeedFollowParams{
			ID:        uuid.New(),
			CreatedAt: t,
			UpdatedAt: t,
			FeedID:    f.ID,
			UserID:    user.ID,
		})
	if err != nil {
		return err
	}

	fmt.Printf("Feed: %s, Current User: %s\n", ff.Feedname, ff.Username)

	return nil
}

func followingFeeds(s *state, cmd command, user database.User) error {
	if len(cmd.args) != 0 {
		return errors.New("no arguments needed")
	}

	ff, err := s.db.GetFeedFollowsForUser(context.Background(), user.ID)
	if err != nil {
		return err
	}

	for _, feed := range ff {
		fmt.Printf("%s\n", feed.Feedname)
	}

	return nil
}

func unfollowFeed(s *state, cmd command, user database.User) error {
	if len(cmd.args) == 0 {
		return errors.New("must specify feed URL")
	}

	fmt.Printf("%v\n", user)
	url := cmd.args[0]
	feed, err := s.db.GetFeedByUrl(context.Background(), url)
	if err != nil {
		return err
	}

	fmt.Printf("%v\n", feed)

	err = s.db.UnfollowFeed(context.Background(),
		database.UnfollowFeedParams{
			FeedID: feed.ID,
			UserID: user.ID,
		})

	if err != nil {
		return err
	}

	fmt.Printf("Unfollowed %s\n", feed.Name)
	return nil
}

func browsePosts(s *state, cmd command, user database.User) error {
	limit := int32(2)
	if len(cmd.args) > 0 {
		userLimit, err := strconv.Atoi(cmd.args[0])
		if err != nil {
			return err
		}

		limit = int32(userLimit)
	}

	fmt.Printf("Retrieving %d posts for user %s(%s)\n", limit, user.Name, user.ID)

	posts, err := s.db.GetPostsForUser(context.Background(),
		database.GetPostsForUserParams{
			UserID: user.ID,
			Limit:  limit,
		})
	if err != nil {
		fmt.Printf("Error getting posts for user\n")
		return err
	}

	fmt.Printf("Retrieved %d posts\n", len(posts))
	for _, post := range posts {
		fmt.Printf("%s\n%v\n%s\n\n", post.Title, post.PublishedAt, post.Description)
	}

	return nil
}
