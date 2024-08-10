package auth

import (
	"fmt"
	"jobList/config"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/sessions"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/discord"
	"github.com/markbates/goth/providers/github"
	"github.com/markbates/goth/providers/google"
)

var (
	SessionName = config.Envs.CookiesAuthSecret
)

type AuthService struct{}

func RequireSession(handlerFunc http.HandlerFunc, auth *AuthService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session, err := auth.GetSessionUser(r)
		if err != nil {
			log.Println("User is not authenticated!")
			http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
			return
		}

		log.Printf("user is authenticated! user: %v!", session.FirstName)

		handlerFunc(w, r)
	}
}

func (s *AuthService) GetSessionUser(r *http.Request) (goth.User, error) {
	session, err := gothic.Store.Get(r, SessionName)
	if err != nil {
		return goth.User{}, err
	}
	u := session.Values["user"]
	if u == nil {
		return goth.User{}, fmt.Errorf("user is not authenticated! %v", u)
	}

	return u.(goth.User), nil
}

// Get a session. We're ignoring the error resulted from decoding an
// existing session: Get() always returns a session, even if empty.
func (s *AuthService) StoreUserSession(w http.ResponseWriter, r *http.Request, user goth.User) error {
	session, _ := gothic.Store.Get(r, SessionName)

	session.Values["user"] = user
	//log.Println(session.Values["user"])
	err := session.Save(r, w)
	if err != nil {
		return err
	}
	log.Println("stored", session.Values["user"])
	return nil
}

func (s *AuthService) RemoveUserSession(w http.ResponseWriter, r *http.Request) {
	session, err := gothic.Store.Get(r, SessionName)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	session.Values["user"] = goth.User{}
	// delete the cookie immediately
	session.Options.MaxAge = -1

	session.Save(r, w)
}

func NewAuthService(store sessions.Store) *AuthService {
	gothic.Store = store
	googleProvider := google.New(
		config.Envs.GoogleKey,
		config.Envs.GoogleSecret,
		buildCallbackURL("google"),
		"profile",
		"email")

	githubProvider := github.New(
		config.Envs.GithubKey,
		config.Envs.GithubSecret,
		buildCallbackURL("github"))

	discordProvider := discord.New(
		config.Envs.DiscordKey,
		config.Envs.DiscordSecret,
		buildCallbackURL("discord"),
		discord.ScopeIdentify,
		discord.ScopeEmail)

	goth.UseProviders(
		githubProvider,
		googleProvider,
		discordProvider,
	)
	return &AuthService{}
}

func buildCallbackURL(provider string) string {
	return fmt.Sprintf("%s/auth/%s/callback", os.Getenv("HOST"), provider)
}

// For google provider removes IDToken and sets FirstName as the user Name for simplicity.
//
// Others providers might do the same with goth user Name
func ParseGothUser(user *goth.User, provider string) {
	switch provider {
	case "google":
		user.Name = user.FirstName
		user.IDToken = ""
	case "discord":
		globalName, ok := user.RawData["global_name"].(string)
		if ok && len(globalName) > 0 {
			user.Name = globalName
		}
	}
}
