package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gofrs/uuid"
	"golang.org/x/oauth2"
)

// tokenSource returns a token source that can be used to refresh the token
func (esi *ESI) tokenSource() (oauth2.TokenSource, error) {
	if esi.token == nil {
		var err error
		esi.token, err = esi.readToken()
		if err != nil {
			return nil, err
		}
	}
	ts := esi.sso.TokenSource(esi.token)
	newToken, err := ts.Token()
	if err != nil {
		log.Println("[ERROR] error getting token")
		return nil, err
	}

	if esi.token != newToken {
		esi.token = newToken
	}

	return ts, nil
}

// /login
func (esi *ESI) handleEsiLogin(w http.ResponseWriter, r *http.Request) {
	state, err := uuid.NewV4()
	if err != nil {
		http.Error(w, "[ERROR] Unable to create random state for auth", http.StatusInternalServerError)
		return
	}

	session, _ := esi.store.Get(r, "session")
	session.Values["state"] = state.String()
	session.Save(r, w)

	// generate SSO URL
	url := esi.sso.AuthorizeURL(state.String(), true, esi.scopes)

	http.Redirect(w, r, url, http.StatusFound)
}

// /callback
func (esi *ESI) handleEsiCallback(w http.ResponseWriter, r *http.Request) {
	code := r.FormValue("code")
	state := r.FormValue("state")
	session, _ := esi.store.Get(r, "session")

	if session.Values["state"] != state {
		http.Error(w, "[ERROR] Bad auth state", http.StatusInternalServerError)
	}

	token, err := esi.sso.TokenExchange(code)
	if err != nil {
		http.Error(w, "Token Exchange Failure", http.StatusInternalServerError)
	}

	// token source refreshes the token in the future
	tokenSource := esi.sso.TokenSource(token)

	// verify
	_, err = esi.sso.Verify(tokenSource)
	if err != nil {
		http.Error(w, "Verify Failure", http.StatusInternalServerError)
	}

	esi.token = token
	esi.storeToken(token)

	fmt.Fprintf(w, "Login Success")
}
