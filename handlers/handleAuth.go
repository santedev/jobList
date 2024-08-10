package handlers

import (
	"context"
	"fmt"
	"jobList/config"
	"jobList/handlers/render"
	"jobList/services/auth"
	a "jobList/views/auth"
	"jobList/views/home"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
)

type AuthHandler struct {
	auth  *auth.AuthService
}

func New(auth *auth.AuthService) *AuthHandler {
	return &AuthHandler{
		auth:  auth,
	}
}

func (h *AuthHandler) HandleProviderLogin(w http.ResponseWriter, r *http.Request) error {
	provider := chi.URLParam(r, "provider")
	ctx := context.WithValue(r.Context(), "provider", provider)
	r = r.WithContext(ctx)

	if checkCompleteUserAuth(r) {
		user, err := gothic.CompleteUserAuth(w, r)
		if err != nil {
			return err
		}
		return render.Template(w, r, home.Index(user))
	}
	user, ok := userFallback(r)
	if ok && user.Provider == provider {
		return render.Template(w, r, home.Index(user))
	}

	gothic.BeginAuthHandler(w, r)
	return nil
}

func (h *AuthHandler) HandleAuthCallback(w http.ResponseWriter, r *http.Request) error {
	provider := chi.URLParam(r, "provider")
	ctx := context.WithValue(r.Context(), "provider", provider)
	r = r.WithContext(ctx)
	user, err := gothic.CompleteUserAuth(w, r)
	if err != nil {
		return err
	}
	auth.ParseGothUser(&user, provider)
	err = h.auth.StoreUserSession(w, r, user)
	if err != nil {
		return err
	}
	w.Header().Set("Location", "/")
	w.WriteHeader(http.StatusTemporaryRedirect)
	return nil
}

func (h *AuthHandler) HandleAuthLogout(w http.ResponseWriter, r *http.Request) error {
	provider := chi.URLParam(r, "provider")
	ctx := context.WithValue(r.Context(), "provider", provider)
	r = r.WithContext(ctx)

	err := gothic.Logout(w, r)
	if err != nil {
		return err
	}

	h.auth.RemoveUserSession(w, r)

	w.Header().Set("Location", "/")
	w.WriteHeader(http.StatusTemporaryRedirect)
	return nil
}

func HandleLogin(w http.ResponseWriter, r *http.Request) error {
	return render.Template(w, r, a.Login())
}

func checkCompleteUserAuth(r *http.Request) bool {
	providerName, err := gothic.GetProviderName(r)
	if err != nil {
		log.Println("provider:", providerName, "err:", err)
		return false
	}
	provider, err := goth.GetProvider(providerName)
	if err != nil {
		log.Println("getProvider err:", err)
		return false
	}
	log.Println("provider from GetProvider:", provider)
	value, err := gothic.GetFromSession(providerName, r)
	if err != nil {
		log.Println("value:", value, "err", err)
		return false
	}
	log.Println("value:", value)
	return true
}

func userFallback(r *http.Request) (goth.User, bool) {
	session, _ := gothic.Store.Get(r, config.Envs.CookiesAuthSecret)

	user, ok := session.Values["user"].(goth.User)
	if ok {
		return user, true
	}
	return goth.User{}, false
}

func getUserFromSession(w http.ResponseWriter, r *http.Request) (goth.User, error) {
	if checkCompleteUserAuth(r) {
		user, err := gothic.CompleteUserAuth(w, r)
		if err != nil {
			return goth.User{}, err
		}
		return user, nil
	}
	user, ok := userFallback(r)
	if !ok {
		return goth.User{}, fmt.Errorf("user fallback for session couldnt work or session doesnt exist")
	}
	return user, nil
}