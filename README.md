# Gator

Gator is a command-line RSS feed aggregator written in Go. It allows users to register, login, add RSS feeds, follow feeds, and scrape posts from followed feeds.

## Installation

1. Ensure you have Go 1.25.4 or later installed.
2. Clone the repository:
   ```
   git clone https://github.com/vasatavit-hub/Gator.git
   cd Gator
   ```
3. Install dependencies:
   ```
   go mod tidy
   ```
4. Build the application:
   ```
   go build -o gator
   ```
5. Set up a PostgreSQL database and configure the connection in the config file (see config section).

## Usage

Run the application with commands:

- `login <username>`: Login as a user
- `register <username>`: Register a new user
- `reset`: Reset the database
- `users`: List all users
- `addfeed <feed_url> <feed_name>`: Add a new RSS feed (requires login)
- `feeds`: List all feeds
- `follow <feed_url>`: Follow a feed (requires login)
- `following`: List feeds you are following (requires login)
- `unfollow <feed_url>`: Unfollow a feed (requires login)
- `scrape <time_between_reqs>`: Scrape posts feeds with specified time between requests

## Configuration

The application uses a configuration file to store database URL and current user. The config is stored in `~/.gatorconfig.json` by default.

## Dependencies

- PostgreSQL database
- Go modules: github.com/google/uuid, github.com/lib/pq
