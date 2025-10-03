package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/francisco/gridironmind/internal/config"
	"github.com/francisco/gridironmind/internal/yahoo"
	"golang.org/x/oauth2"
)

func main() {
	// Load config
	cfg := config.LoadConfig()

	if cfg.YahooClientID == "" || cfg.YahooClientSecret == "" {
		log.Fatal("YAHOO_CLIENT_ID and YAHOO_CLIENT_SECRET must be set")
	}

	// Create OAuth config
	oauthConfig := &oauth2.Config{
		ClientID:     cfg.YahooClientID,
		ClientSecret: cfg.YahooClientSecret,
		RedirectURL:  "oob", // Out-of-band for command-line apps
		Scopes:       []string{},
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://api.login.yahoo.com/oauth2/request_auth",
			TokenURL: "https://api.login.yahoo.com/oauth2/get_token",
		},
	}

	fmt.Println("=== Yahoo Fantasy Sports OAuth Setup ===")
	fmt.Println()

	// Start callback server
	callbackChan := make(chan string)
	srv := startCallbackServer(callbackChan)
	defer srv.Shutdown(context.Background())

	// Generate auth URL with local callback
	oauthConfig.RedirectURL = "http://localhost:8888/callback"
	authURL := oauthConfig.AuthCodeURL("state", oauth2.AccessTypeOffline)

	fmt.Println("Step 1: Open this URL in your browser:")
	fmt.Println()
	fmt.Println(authURL)
	fmt.Println()
	fmt.Println("Step 2: Authorize the app and you'll be redirected back automatically...")
	fmt.Println()

	// Wait for callback
	code := <-callbackChan
	fmt.Println("✓ Received authorization code")
	fmt.Println()

	// Exchange code for token
	ctx := context.Background()
	token, err := oauthConfig.Exchange(ctx, code)
	if err != nil {
		log.Fatalf("Failed to exchange code for token: %v", err)
	}

	fmt.Println("✓ Successfully obtained OAuth tokens!")
	fmt.Println()
	fmt.Println("=== Token Details ===")
	fmt.Printf("Access Token: %s\n", token.AccessToken)
	fmt.Printf("Refresh Token: %s\n", token.RefreshToken)
	fmt.Printf("Token Type: %s\n", token.TokenType)
	fmt.Printf("Expires At: %v\n", token.Expiry)
	fmt.Println()

	// Test the token
	fmt.Println("Step 3: Testing the token...")
	client := yahoo.NewClientWithToken(cfg.YahooClientID, cfg.YahooClientSecret, token)

	// Try to fetch player rankings
	rankings, err := client.FetchPlayerRankings(ctx, "QB", 1)
	if err != nil {
		log.Printf("Warning: Failed to fetch rankings (might need league context): %v", err)
	} else if rankings != nil {
		fmt.Printf("✓ Successfully fetched %d QB rankings!\n", len(rankings.Players))
	}
	fmt.Println()

	// Print instructions for setting up Heroku
	fmt.Println("=== Next Steps ===")
	fmt.Println()
	fmt.Println("Run this command to set the refresh token on Heroku:")
	fmt.Println()
	fmt.Printf("heroku config:set YAHOO_REFRESH_TOKEN=\"%s\"\n", token.RefreshToken)
	fmt.Println()
	fmt.Println("Or add to your .env file:")
	fmt.Printf("YAHOO_REFRESH_TOKEN=%s\n", token.RefreshToken)
	fmt.Println()
}

func startCallbackServer(callbackChan chan string) *http.Server {
	mux := http.NewServeMux()

	mux.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")
		if code == "" {
			http.Error(w, "No code in callback", http.StatusBadRequest)
			return
		}

		callbackChan <- code

		w.Header().Set("Content-Type", "text/html")
		fmt.Fprintf(w, `
			<html>
			<head><title>Yahoo OAuth Success</title></head>
			<body style="font-family: Arial; text-align: center; padding: 50px;">
				<h1 style="color: #5f01d1;">✓ Authorization Successful!</h1>
				<p>You can close this window and return to the terminal.</p>
			</body>
			</html>
		`)
	})

	srv := &http.Server{
		Addr:    ":8888",
		Handler: mux,
	}

	go func() {
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			log.Printf("Callback server error: %v", err)
		}
	}()

	// Give server time to start
	time.Sleep(100 * time.Millisecond)

	return srv
}
