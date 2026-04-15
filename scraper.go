package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/KJBrock/bootdev_gator/internal/database"
	"github.com/google/uuid"
)

func scrapeFeeds(s *state) error {

	nextFeed, err := s.db.GetNextFeedToFetch(context.Background())
	if err != nil {
		return err
	}

	fmt.Printf("---> Fetching %s\n", nextFeed.Name)
	err = s.db.MarkFeedFetched(context.Background(), nextFeed.ID)
	if err != nil {
		return err
	}

	rssFeed, err := fetchFeed(context.Background(), nextFeed.Url)
	if err != nil {
		fmt.Printf("error fetching feed: %v\n", err)
		return err
	}

	fmt.Printf("================\n%s\n================\n", rssFeed.Channel.Title)

	for _, item := range rssFeed.Channel.Item {
		// fmt.Printf("%d: %v\n", i, item)

		t := time.Now()
		published, err := time.Parse(time.RFC1123Z, item.PubDate)
		if err != nil {
			fmt.Printf("Time parse error: %s\n", item.PubDate)
			published = t
		}

		_, err = s.db.CreatePost(context.Background(),
			database.CreatePostParams{
				ID:          uuid.New(),
				CreatedAt:   t,
				UpdatedAt:   t,
				PublishedAt: published,
				Title:       item.Title,
				Url:         item.Link,
				Description: item.Description,
				FeedID:      nextFeed.ID,
			})
		if err != nil {
			// We expect lots of duplicates for this.
			if strings.Contains(err.Error(), "posts_url_key") && strings.Contains(err.Error(), "(23505)") {
				continue
			}

			fmt.Printf("Error creating post: %v\n", err)
			return err
		}

		// fmt.Printf("%v\n", post)
	}

	return nil
}
