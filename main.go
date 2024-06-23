package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	oauth2api "google.golang.org/api/oauth2/v2"
	"google.golang.org/api/option"
)

var (
    googleOauthConfig = &oauth2.Config{}
    oauthStateString = ""
)

func main() {
    envLoad()
    oauthInit();
    r := gin.Default()

    r.GET("/", func(c *gin.Context) {
        var html = `<html><body><a href="/login">Google Log In</a></body></html>`
        c.Writer.WriteHeader(http.StatusOK)
        c.Writer.Write([]byte(html))
    })

    r.GET("/login", handleGoogleLogin)
    r.GET("/auth/callback", handleGoogleCallback)

    r.Run(":8080")
}

func envLoad() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Error loading env target")
	}
}

func oauthInit() {
    googleOauthConfig = &oauth2.Config{
        RedirectURL:  "http://localhost:8080/auth/callback",
        ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
        ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
        Scopes: []string{
            "https://www.googleapis.com/auth/userinfo.profile",
            "https://www.googleapis.com/auth/userinfo.email",
        },
        Endpoint: google.Endpoint,
    }
    oauthStateString = os.Getenv("GOOGLE_OAUTH_STATE")
}

func handleGoogleLogin(c *gin.Context) {
    url := googleOauthConfig.AuthCodeURL(oauthStateString)
    c.Redirect(http.StatusTemporaryRedirect, url)
}

func handleGoogleCallback(c *gin.Context) {
    state := c.Query("state")
    if state != oauthStateString {
        log.Println("invalid oauth state")
        c.Redirect(http.StatusTemporaryRedirect, "/")
        return
    }

    code := c.Query("code")
    token, err := googleOauthConfig.Exchange(context.Background(), code)
    if err != nil {
        log.Printf("code exchange failed: %s", err.Error())
        c.Redirect(http.StatusTemporaryRedirect, "/")
        return
    }

    oauth2Service, err := oauth2api.NewService(context.Background(), option.WithTokenSource(googleOauthConfig.TokenSource(context.Background(), token)))
        if err != nil {
        log.Printf("failed to create oauth2 service: %s", err.Error())
        c.Redirect(http.StatusTemporaryRedirect, "/")
        return
    }

    userInfo, err := oauth2Service.Userinfo.Get().Do()
    if err != nil {
        log.Printf("failed to get user info: %s", err.Error())
        c.Redirect(http.StatusTemporaryRedirect, "/")
        return
    }

    var html string
    html += "<li>Name: " + userInfo.Name + "</li>"
    html += "<li>Email: " + userInfo.Email + "</li>"
    html += "<li>ID: " + userInfo.Id + "</li>"
    html += "<li>Picture: " + userInfo.Picture + "</li>aaaa"
    
    html = "<html><body><ul>" + html + "</ul></body></html>"
    c.Writer.WriteHeader(http.StatusOK)
    c.Writer.Write([]byte(html))
}
