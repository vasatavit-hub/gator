package main

import (
	"context"
	"database/sql"
	"encoding/xml"
	"fmt"
	"html"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/vasatavit-hub/Gator/internal/database"
)

type command struct {
	name    string
	command []string
}

type commands struct {
	commandMap map[string]func(*state, command) error
}

type RSSFeed struct {
	Channel struct {
		Title       string    `xml:"title"`
		Link        string    `xml:"link"`
		Description string    `xml:"description"`
		Item        []RSSItem `xml:"item"`
	} `xml:"channel"`
}

type RSSItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
}

func (c *commands) run(s *state, cmd command) error {
	err := c.commandMap[cmd.name](s, cmd)
	if err != nil {
		return fmt.Errorf("Failed to run command %s: %v", cmd.name, err)
	}
	return nil
}

func (c *commands) register(name string, f func(*state, command) error) {
	if c.commandMap == nil {

	}
	c.commandMap[name] = f
}

func handlerLogin(s *state, cmd command) error {
	if len(cmd.command) < 1 {
		return fmt.Errorf("The login handler expects a single argument, the username")
	}
	// Handle login logic here
	_, err := s.db.GetUser(context.Background(), cmd.command[0])
	if err != nil {
		fmt.Printf("User %s does not exist", cmd.command[0])
		os.Exit(1)
	}

	s.cfg.SetUser(cmd.command[0])
	fmt.Printf("User %s logged in successfully\n", cmd.command[0])

	return nil
}

func handlerRegister(s *state, cmd command) error {
	if len(cmd.command) < 1 {
		return fmt.Errorf("The register handler expects a single argument, the username")
	}
	// Handle registration logic here
	// Check if user already exists

	_, err := s.db.GetUser(context.Background(), cmd.command[0])
	if err == nil {
		fmt.Printf("User %s already exists", cmd.command[0])
		os.Exit(1)
	}

	// Create new user

	id := uuid.New()
	fmt.Printf("Generated user ID: %s\n", id)

	newUser, err := s.db.CreateUser(context.Background(), database.CreateUserParams{
		ID:        id,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      cmd.command[0],
	})
	if err != nil {
		return fmt.Errorf("Failed to create user: %v", err)
	}

	s.cfg.SetUser(cmd.command[0])

	fmt.Printf("User %s registered successfully\n", newUser.Name)
	fmt.Printf("ID: %v\n", newUser.ID)
	fmt.Printf("CreatedAt: %v\n", newUser.CreatedAt)
	fmt.Printf("UpdatedAt: %v\n", newUser.UpdatedAt)

	return nil
}

func handlerReset(s *state, cmd command) error {
	// Handle reset logic here
	err := s.db.CleanDatabase(context.Background())
	if err != nil {
		log.Fatalf("Failed to reset database: %v", err)
	}
	fmt.Println("Database reset successfully")
	return nil
}

func handlerGetUsers(s *state, cmd command) error {
	// Handle get-users logic here
	users, err := s.db.GetUsers(context.Background())
	if err != nil {
		log.Fatalf("Failed to get users: %v", err)
	}

	current := s.cfg.Current_user_name

	for _, user := range users {
		fmt.Printf("Name: %v", user.Name)
		if user.Name == current {
			fmt.Printf(" (current)")
		}
		fmt.Printf("\n")
	}
	return nil
}

func fetchFeed(ctx context.Context, feedURL string) (*RSSFeed, error) {
	httpClient := &http.Client{
		Timeout: 10 * time.Second,
	}

	req, err := http.NewRequestWithContext(ctx, "GET", feedURL, nil)
	if err != nil {
		return nil, fmt.Errorf("Failed to create HTTP request: %v", err)
	}

	req.Header.Set("User-Agent", "gator")

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Failed to fetch RSS feed: %v", err)
	}
	defer resp.Body.Close()

	var feed RSSFeed
	body, _ := io.ReadAll(resp.Body)
	xml.Unmarshal(body, &feed)

	feed.Channel.Title = html.UnescapeString(feed.Channel.Title)
	feed.Channel.Description = html.UnescapeString(feed.Channel.Description)

	for i := range feed.Channel.Item {
		feed.Channel.Item[i].Title = html.UnescapeString(feed.Channel.Item[i].Title)
		feed.Channel.Item[i].Description = html.UnescapeString(feed.Channel.Item[i].Description)
	}

	return &feed, nil

}

func handlerAgg(s *state, cmd command) error {

	URL := "https://techcrunch.com/feed/"

	// Do something with the fetched feed, e.g., display it or save it
	//fmt.Printf("Fetching feed with URL %s\n", URL)
	resp, err := fetchFeed(context.Background(), URL)
	if err != nil {
		return fmt.Errorf("Failed to fetch feed: %v", err)
	}

	fmt.Printf("Feed Title: %s\n", resp.Channel.Title)
	//fmt.Printf("Feed Description: %s\n", resp.Channel.Description)
	for _, item := range resp.Channel.Item {
		fmt.Printf("Item Title: %s\n", item.Title)
		//fmt.Printf("Item Link: %s\n", item.Link)
		//fmt.Printf("Item Description: %s\n", item.Description)
		//fmt.Printf("Item Publication Date: %s\n", item.PubDate)
		fmt.Println()
	}

	return nil
}

func addfeed(name, url string, s *state, user database.GetUserRow) (*database.Feed, error) {
	// Handle add-feed logic here
	id := uuid.New()

	// Get current user ID

	currentUserID := user.ID

	// Create new feed

	newfeed, err := s.db.CreateFeed(context.Background(), database.CreateFeedParams{
		ID:        id,
		Name:      name,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Url:       url,
		UserID:    currentUserID,
	})

	return &newfeed, err
}

func handlerAddFeed(s *state, cmd command, user database.GetUserRow) error {
	if len(cmd.command) < 2 {
		return fmt.Errorf("The add-feed handler expects two arguments, the feed name and the feed URL")
	}

	name := cmd.command[0]
	url := cmd.command[1]

	feed, err := addfeed(name, url, s, user)
	if err != nil {
		return fmt.Errorf("Failed to create feed: %v", err)
	}

	fmt.Printf("Feed %s added successfully\n", feed.Name)
	fmt.Printf("ID: %v\n", feed.ID)
	fmt.Printf("CreatedAt: %v\n", feed.CreatedAt)
	fmt.Printf("UpdatedAt: %v\n", feed.UpdatedAt)
	fmt.Printf("URL: %v\n", feed.Url)

	_, err = follow(feed.Url, s, user)
	if err != nil {
		return fmt.Errorf("Failed to follow feed: %v", err)
	}

	return nil
}

func listFeeds(s *state) (*[]database.ListFeedsRow, error) {
	feeds, err := s.db.ListFeeds(context.Background())
	return &feeds, err
}

func handlerListFeeds(s *state, cmd command) error {
	feeds, err := listFeeds(s)
	if err != nil {
		return fmt.Errorf("Failed to list feeds: %v", err)
	}

	for _, feed := range *feeds {
		fmt.Printf("Feed Name: %s\n", feed.Name)
		fmt.Printf("Feed URL: %s\n", feed.Url)
		fmt.Printf("User Name: %s\n", feed.Name_2)
		fmt.Println()
	}
	return nil
}

func follow(ulr string, s *state, user database.GetUserRow) ([]database.CreateFeedFollowRow, error) {
	// Handle follow logic here
	feeds, err := s.db.ListFeeds(context.Background())
	if err != nil {
		return nil, fmt.Errorf("Failed to list feeds: %v", err)
	}

	// Get current user ID
	currentUserID := user.ID

	for _, feed := range feeds {
		if feed.Url == ulr {
			fmt.Printf("Found feed with URL %s\n", ulr)

			id := uuid.New()

			// Create new feed follow

			newFeedFollow, err := s.db.CreateFeedFollow(context.Background(), database.CreateFeedFollowParams{
				ID:        id,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
				UserID:    currentUserID,
				FeedID:    feed.ID,
			})
			return newFeedFollow, err
		}
	}

	return nil, fmt.Errorf("No feed found with URL %s", ulr)

}

func handlerFollow(s *state, cmd command, user database.GetUserRow) error {
	if len(cmd.command) < 1 {
		return fmt.Errorf("The follow handler expects one argument, the feed URL")
	}

	url := cmd.command[0]

	feed, err := follow(url, s, user)
	if err != nil {
		return fmt.Errorf("Failed to create feed follow: %v", err)
	}

	fmt.Printf("Followed feed successfully\n")
	fmt.Printf("Feed Name: %v\n", feed[0].FeedName)
	fmt.Printf("User Name: %v\n", feed[0].UserName)

	return nil
}

func listFeedFollows(s *state, user database.GetUserRow) (*[]database.GetFeedFollowsForUserRow, error) {
	// Handle list feed follows logic here
	// Get current user ID

	currentUserID := user.ID

	follows, err := s.db.GetFeedFollowsForUser(context.Background(), currentUserID)
	return &follows, err
}

func handlerListFeedFollows(s *state, cmd command, user database.GetUserRow) error {
	follows, err := listFeedFollows(s, user)
	if err != nil {
		return fmt.Errorf("Failed to list feed follows: %v", err)
	}

	for _, follow := range *follows {
		fmt.Printf("Feed Name: %s\n", follow.FeedName)
		fmt.Printf("User Name: %s\n", follow.UserName)
		fmt.Printf("Feeds ID: %s\n", follow.FeedID)
		fmt.Printf("User ID: %s\n", user.ID)
	}
	return nil
}

func middlewareLoggedIn(handler func(s *state, cmd command, user database.GetUserRow) error) func(*state, command) error {
	return func(s *state, cmd command) error {
		currentUserName := s.cfg.Current_user_name
		if currentUserName == "" {
			return fmt.Errorf("You must be logged in to run this command")
		}

		user, err := s.db.GetUser(context.Background(), currentUserName)
		if err != nil {
			return fmt.Errorf("Failed to get current user: %v", err)
		}

		return handler(s, cmd, user)
	}
}

func unfollow(s *state, feedFollowID uuid.UUID, user database.GetUserRow) error {
	// Handle delete feed follow logic here
	currentUserID := user.ID

	fmt.Printf("Unfollowing feed with follow ID %s for user ID %s\n", feedFollowID, currentUserID)

	_, err := s.db.DeleteFeedFollow(context.Background(), database.DeleteFeedFollowParams{
		FeedID: feedFollowID,
		UserID: currentUserID,
	})
	return err
}

func handlerUnfollow(s *state, cmd command, user database.GetUserRow) error {
	if len(cmd.command) < 1 {
		return fmt.Errorf("The unfollow handler expects one argument, the feed URL")
	}

	feedFollowID, err := feedID(s, cmd.command[0])
	if err != nil {
		return fmt.Errorf("Failed to get feed ID: %v", err)
	}

	err = unfollow(s, *feedFollowID, user)
	if err != nil {
		return fmt.Errorf("Failed to delete feed follow: %v", err)
	}

	fmt.Printf("Unfollowed feed successfully\n")

	return nil

}

func feedID(s *state, feedURL string) (*uuid.UUID, error) {
	feeds, err := s.db.ListFeeds(context.Background())
	if err != nil {
		return nil, fmt.Errorf("Failed to list feeds: %v", err)
	}

	for _, feed := range feeds {
		if feed.Url == feedURL {
			return &feed.ID, nil
		}
	}

	return nil, fmt.Errorf("No feed found with URL %s", feedURL)
}

func scrapeFeeds(s *state) error {
	// Get next feed to fetch
	feeds, err := s.db.GetNextFeedToFetch(context.Background())
	if err != nil {
		return fmt.Errorf("Failed to get next feed to fetch: %v", err)
	}
	// If no feeds to fetch, return
	if len(feeds) == 0 {
		fmt.Println("No feeds to fetch")
		return nil
	}

	for _, feed := range feeds {
		fmt.Printf("Next feed to fetch: %s with URL %s\n", feed.Name, feed.Url)
		// Mark feed as fetched
		err = s.db.MarkFeedFetched(context.Background(), feed.ID)
		if err != nil {
			log.Fatalf("Failed to mark feed as fetched: %v", err)
		}

		// Fetch feed
		fmt.Printf("Fetching feed %s with URL %s\n", feed.Name, feed.Url)
		resp, err := fetchFeed(context.Background(), feed.Url)
		if err != nil {
			log.Fatalf("Failed to fetch feed: %v", err)
		}

		//

		fmt.Printf("Feed Title: %s\n", resp.Channel.Title)
		fmt.Printf("Feed Description: %s\n", resp.Channel.Description)
		for i, item := range resp.Channel.Item {
			_, err := createPost(s, feed.ID, item)
			if err != nil {
				if strings.Contains(err.Error(), "duplicate key") && strings.Contains(err.Error(), "posts_url_key") {
					continue
				} else {
					log.Printf("Failed to create post for item %d: %v", i, err)
				}

			}
		}
	}
	return nil
}

func handlerScrapeFeeds(s *state, cmd command) error {
	if len(cmd.command) < 1 {
		return fmt.Errorf("The scrape handler expects one argument, the scrape interval")
	}

	scrapeInterval, err := time.ParseDuration(cmd.command[0])
	if err != nil {
		return fmt.Errorf("Failed to parse scrape interval: %v", err)
	}
	ticker := time.NewTicker(scrapeInterval)
	defer ticker.Stop()

	for ; ; <-ticker.C {
		err := scrapeFeeds(s)
		if err != nil {
			log.Printf("Failed to scrape feeds: %v", err)
		}
	}
}

func createPost(s *state, feedID uuid.UUID, item RSSItem) (*database.Post, error) {
	id := uuid.New()

	publishedAt, err := time.Parse(time.RFC1123Z, item.PubDate)
	if err != nil {
		return nil, fmt.Errorf("Failed to parse publication date: %v", err)
	}

	newPost, err := s.db.CreatePost(context.Background(), database.CreatePostParams{
		ID:        id,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Title:     item.Title,
		Url:       item.Link,
		Description: sql.NullString{
			String: item.Description,
			Valid:  true,
		},
		PublishedAt: sql.NullTime{
			Time:  publishedAt,
			Valid: true,
		},
		FeedID: feedID,
	})

	return &newPost, err
}

func handlerBrowsePosts(s *state, cmd command, user database.GetUserRow) error {
	var limit int64
	var err error

	if len(cmd.command) >= 1 {
		limit, err = strconv.ParseInt(cmd.command[0], 10, 32)
		if err != nil {
			limit = 2
		}
		if limit != int64(int32(limit)) {
			limit = 2
		}

	} else {
		limit = 2
	}

	posts, err := browsePosts(s, user, int32(limit))
	if err != nil {
		return fmt.Errorf("Failed to browse posts: %v", err)
	}

	for _, post := range posts {
		fmt.Printf("Title: %s\n", post.Title)
		fmt.Printf("URL: %s\n", post.Url)
		fmt.Printf("Description: %s\n", post.Description.String)
		fmt.Printf("Published At: %s\n", post.PublishedAt.Time)
		fmt.Printf("Feed Name: %s\n", post.FeedName)
		fmt.Println()
	}

	return nil
}

func browsePosts(s *state, user database.GetUserRow, limit int32) ([]database.GetPostsForUserRow, error) {
	posts, err := s.db.GetPostsForUser(context.Background(), database.GetPostsForUserParams{UserID: user.ID, Limit: limit})
	return posts, err
}
