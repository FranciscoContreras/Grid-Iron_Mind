package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"golang.org/x/oauth2"
)

var (
	oauthConfig *oauth2.Config
	state       = "grid-iron-mind-yahoo-oauth"
)

func main() {
	clientID := os.Getenv("YAHOO_CLIENT_ID")
	clientSecret := os.Getenv("YAHOO_CLIENT_SECRET")
	redirectURL := os.Getenv("YAHOO_REDIRECT_URL")

	if clientID == "" || clientSecret == "" {
		log.Fatal("YAHOO_CLIENT_ID and YAHOO_CLIENT_SECRET must be set")
	}

	if redirectURL == "" {
		redirectURL = "https://grid-iron-mind-71cc9734eaf4.herokuapp.com/yahoo/callback"
	}

	oauthConfig = &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURL,
		Scopes:       []string{},
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://api.login.yahoo.com/oauth2/request_auth",
			TokenURL: "https://api.login.yahoo.com/oauth2/get_token",
		},
	}

	http.HandleFunc("/yahoo/auth", handleYahooAuth)
	http.HandleFunc("/yahoo/callback", handleYahooCallback)
	http.HandleFunc("/", handleHome)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("OAuth Helper running on port %s", port)
	log.Printf("Visit: http://localhost:%s/ to start OAuth flow", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func handleHome(w http.ResponseWriter, r *http.Request) {
	html := `
<!DOCTYPE html>
<html>
<head>
    <title>Yahoo OAuth Setup - Grid Iron Mind</title>
    <style>
        body {
            font-family: Arial, sans-serif;
            max-width: 800px;
            margin: 50px auto;
            padding: 20px;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: white;
        }
        .container {
            background: rgba(255, 255, 255, 0.1);
            padding: 40px;
            border-radius: 15px;
            backdrop-filter: blur(10px);
        }
        h1 {
            margin-top: 0;
        }
        .btn {
            display: inline-block;
            padding: 15px 30px;
            background: #fff;
            color: #764ba2;
            text-decoration: none;
            border-radius: 8px;
            font-weight: bold;
            margin-top: 20px;
            transition: transform 0.2s;
        }
        .btn:hover {
            transform: scale(1.05);
        }
        code {
            background: rgba(0,0,0,0.3);
            padding: 2px 6px;
            border-radius: 3px;
        }
    </style>
</head>
<body>
    <div class="container">
        <h1>🏈 Yahoo Fantasy Sports OAuth Setup</h1>
        <p>This tool will help you authorize Grid Iron Mind to access Yahoo Fantasy Sports API.</p>

        <h3>Steps:</h3>
        <ol>
            <li>Click the button below to start the OAuth flow</li>
            <li>Log in with your Yahoo account</li>
            <li>Authorize the application</li>
            <li>You'll be redirected back with your tokens</li>
        </ol>

        <a href="/yahoo/auth" class="btn">🔐 Start OAuth Flow</a>

        <p style="margin-top: 40px; font-size: 12px; opacity: 0.8;">
            Make sure you have YAHOO_CLIENT_ID and YAHOO_CLIENT_SECRET set in your environment.
        </p>
    </div>
</body>
</html>
`
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprint(w, html)
}

func handleYahooAuth(w http.ResponseWriter, r *http.Request) {
	url := oauthConfig.AuthCodeURL(state, oauth2.AccessTypeOffline)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func handleYahooCallback(w http.ResponseWriter, r *http.Request) {
	if r.URL.Query().Get("state") != state {
		http.Error(w, "State mismatch", http.StatusBadRequest)
		return
	}

	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "No code in callback", http.StatusBadRequest)
		return
	}

	// Exchange code for token
	ctx := context.Background()
	token, err := oauthConfig.Exchange(ctx, code)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to exchange token: %v", err), http.StatusInternalServerError)
		return
	}

	// Display success page with token
	html := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <title>OAuth Success - Grid Iron Mind</title>
    <style>
        body {
            font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif;
            max-width: 900px;
            margin: 50px auto;
            padding: 20px;
            background: linear-gradient(135deg, #667eea 0%%, #764ba2 100%%);
            color: white;
        }
        .container {
            background: rgba(255, 255, 255, 0.1);
            padding: 40px;
            border-radius: 15px;
            backdrop-filter: blur(10px);
        }
        .success {
            font-size: 48px;
            text-align: center;
            margin-bottom: 20px;
        }
        h1 {
            text-align: center;
            margin-top: 0;
        }
        .token-box {
            background: rgba(0,0,0,0.3);
            padding: 20px;
            border-radius: 8px;
            margin: 20px 0;
            word-wrap: break-word;
            font-family: monospace;
            font-size: 13px;
        }
        .label {
            font-weight: bold;
            color: #ffd700;
            margin-top: 15px;
            display: block;
        }
        .command {
            background: rgba(0,0,0,0.5);
            padding: 15px;
            border-radius: 5px;
            margin: 10px 0;
            border-left: 4px solid #ffd700;
        }
        .copy-btn {
            background: #ffd700;
            color: #764ba2;
            border: none;
            padding: 8px 15px;
            border-radius: 5px;
            cursor: pointer;
            font-weight: bold;
            margin-left: 10px;
        }
        .copy-btn:hover {
            background: #ffed4e;
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="success">✅</div>
        <h1>OAuth Authorization Successful!</h1>

        <p><strong>Your Yahoo Fantasy Sports API is now connected.</strong></p>

        <div class="token-box">
            <span class="label">Refresh Token:</span>
            <div id="refresh-token">%s</div>

            <span class="label">Access Token:</span>
            <div id="access-token">%s</div>

            <span class="label">Expires At:</span>
            <div>%s</div>
        </div>

        <h3>🚀 Next Steps:</h3>

        <p>1. Set the refresh token on Heroku by running this command in your terminal:</p>
        <div class="command">
            <code id="heroku-cmd">heroku config:set YAHOO_REFRESH_TOKEN="%s"</code>
            <button class="copy-btn" onclick="copyToClipboard('heroku-cmd')">Copy</button>
        </div>

        <p>2. Or add to your local <code>.env</code> file:</p>
        <div class="command">
            <code id="env-var">YAHOO_REFRESH_TOKEN=%s</code>
            <button class="copy-btn" onclick="copyToClipboard('env-var')">Copy</button>
        </div>

        <p>3. Restart your application to pick up the new token.</p>

        <p style="margin-top: 40px; padding: 20px; background: rgba(255,255,0,0.1); border-radius: 8px;">
            <strong>⚠️ Security Note:</strong> Keep your refresh token secret!
            It allows access to your Yahoo Fantasy data. Never commit it to version control.
        </p>
    </div>

    <script>
        function copyToClipboard(elementId) {
            const element = document.getElementById(elementId);
            const text = element.innerText;
            navigator.clipboard.writeText(text).then(() => {
                alert('Copied to clipboard!');
            });
        }
    </script>
</body>
</html>
`, token.RefreshToken, token.AccessToken, token.Expiry.String(), token.RefreshToken, token.RefreshToken)

	w.Header().Set("Content-Type", "text/html")
	fmt.Fprint(w, html)

	// Also log to server console
	log.Println("=== Yahoo OAuth Success ===")
	log.Printf("Refresh Token: %s", token.RefreshToken)
	log.Printf("Access Token: %s", token.AccessToken)
	log.Printf("Expires: %v", token.Expiry)
	log.Println("============================")
}
