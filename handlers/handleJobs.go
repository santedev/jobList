package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"jobList/handlers/render"
	"jobList/scrape"
	t "jobList/scrape/types"
	"jobList/store"
	"jobList/views/components"
	"jobList/views/savedJobs"
	"log"

	"net/http"
	"strings"
)

func HandleJobs(w http.ResponseWriter, r *http.Request) error {
	sites, siteList, err := parseReq(r)
	if err != nil {
		return err
	}
	user, err := getUserFromSession(w, r)
	if err != nil {
		log.Println(err.Error())
	}
	err = scrape.GetJobs(w, r, sites, siteList, user)
	if err != nil {
		return err
	}
	return nil
}

func HandleGetSavedJobs(w http.ResponseWriter, r *http.Request) error {
	user, err := getUserFromSession(w, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return err
	}
	if len(user.UserID) <= 0 {
		http.Error(w, "userID is void string", http.StatusBadRequest)
		return fmt.Errorf("userID is void string")
	}
	return scrape.GetSavedJobs(w, r, user)
}

func HandleSavedJobs(w http.ResponseWriter, r *http.Request) error {
	user, err := getUserFromSession(w, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return err
	}
	if len(user.UserID) <= 0 {
		http.Error(w, "userID is void string", http.StatusBadRequest)
		return fmt.Errorf("userID is void string")
	}

	return render.Template(w, r, savedJobs.Index(user))
}

func HandleSaveJobs(w http.ResponseWriter, r *http.Request) error {
	user, err := getUserFromSession(w, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return err
	}
	err = r.ParseForm()
	log.Println(r.Form)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return err
	}
	jbLink := r.Form.Get("jobLink")
	if len(jbLink) <= 0 {
		http.Error(w, "jobLink is void string", http.StatusBadRequest)
		return fmt.Errorf("jobLink from form is null")
	}
	if len(user.UserID) <= 0 {
		http.Error(w, "userID is void string", http.StatusBadRequest)
		return fmt.Errorf("uid is null")
	}
	err = store.DB.SaveJobOffer(user.UserID, jbLink)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}
	return render.Template(w, r, components.Favorite(t.JobStrct{Saved: "saved", JobLink: jbLink}))
}

func HandleUnsaveJobs(w http.ResponseWriter, r *http.Request) error {
	user, err := getUserFromSession(w, r)
	if err != nil {
		return err
	}
	err = r.ParseForm()
	log.Println(r.Form)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return err
	}
	jbLink := r.Form.Get("jobLink")
	log.Println(jbLink)
	if len(jbLink) <= 0 {
		http.Error(w, "jobLink is void string", http.StatusBadRequest)
		return fmt.Errorf("jobLink from form is null")
	}
	if len(user.UserID) <= 0 {
		http.Error(w, "userID is void string", http.StatusBadRequest)
		return fmt.Errorf("uid is null")
	}
	err = store.DB.UnsaveJobOffer(user.UserID, jbLink)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}
	return render.Template(w, r, components.Favorite(t.JobStrct{Saved: "unsaved", JobLink: jbLink}))
}

func parseReq(r *http.Request) (t.SitesStrct, []string, error) {
	b, err := io.ReadAll(r.Body)
	if err != nil {
		return t.SitesStrct{}, []string{}, err
	}
	fmt.Println(string(b))
	defer r.Body.Close()
	var dj t.MainSitesStrct
	if err := json.Unmarshal(b, &dj); err != nil {
		return t.SitesStrct{}, nil, err
	}
	sites := dj.Data
	siteStr := dj.Sites
	if siteStr == "" {
		return t.SitesStrct{}, []string{}, fmt.Errorf("no 'sites' field found in form data")
	}

	var siteList []string
	if len(siteStr) > 0 {
		siteList = strings.Split(siteStr, ",")
	}
	return sites, siteList, nil
}
