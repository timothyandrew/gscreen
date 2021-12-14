package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"os/user"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

func initialAuth(config oauth2.Config) *oauth2.Token {
	url := config.AuthCodeURL("state", oauth2.AccessTypeOffline, oauth2.ApprovalForce)
	fmt.Printf("Visit the URL for the auth dialog: %v\n	", url)

	buf := bufio.NewReader(os.Stdin)
	fmt.Print("> ")
	authCode, err := buf.ReadBytes('\n')
	if err != nil {
		log.Fatalf("Failed to get authCode", err)
	}
	authCodeString := strings.TrimSpace(string(authCode))

	print("code: ", authCodeString)

	// Handle the exchange code to initiate a transport.
	tok, err := config.Exchange(context.Background(), authCodeString)
	if err != nil {
		log.Fatal(err)
	}

	return tok
}

func main() {
	ctx := context.Background()

	usr, _ := user.Current()
	err := godotenv.Load(fmt.Sprintf("%s/.gscreen", usr.HomeDir))
	if err != nil {
		log.Println("~/.gscreen not present")
	}

	config := &oauth2.Config{
		ClientID:     os.Getenv("CLIENT_ID"),
		ClientSecret: os.Getenv("CLIENT_SECRET"),
		RedirectURL:  "http://www.example.com/oauth2callback",
		Scopes:       []string{"https://www.googleapis.com/auth/photoslibrary.readonly"},
		Endpoint:     google.Endpoint,
	}

	refreshToken := os.Getenv("REFRESH_TOKEN")
	var token *oauth2.Token

	if refreshToken == "" {
		token = initialAuth(*config)
	} else {
		token = new(oauth2.Token)
		token.AccessToken = "FOO"
		token.Expiry = time.Now().Add(-1 * time.Hour)
		token.RefreshToken = refreshToken
		token.TokenType = "Bearer"
	}

	client := config.Client(ctx, token)
	cache := InitializeMetadataCache(ctx, client)
	InitializeHttpServer(cache)
}
