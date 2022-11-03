package handlers

import (
	"fmt"
	"log"
	"net/http"
	"twidel/service"

	"github.com/kkdai/twitter"
)

const (
	//CallbackURL: This URL need note as follow:
	// 1. Could not be localhost, change your hosts to a specific domain name
	// 2. This setting must be identical with your app setting on twitter Dev
	CallbackURL string = "http://bostan.com:9090/maketoken"
)

type Handlers struct {
	Ping             http.HandlerFunc
	Index            http.HandlerFunc
	Timeline         http.HandlerFunc
	InitTwitterLogin http.HandlerFunc
	GetTwitterToken  http.HandlerFunc
	GetHomeTimeLine  http.HandlerFunc
	GetUserTimeLine  http.HandlerFunc
}

func NewHandlers(s *service.Service) Handlers {
	return Handlers{
		Ping: func(w http.ResponseWriter, r *http.Request) {

		},

		Index: func(w http.ResponseWriter, r *http.Request) {
			index(w, r, s)
		},

		InitTwitterLogin: func(w http.ResponseWriter, r *http.Request) {
			InitTwitterLogin(w, r, s)
		},
		GetTwitterToken: func(w http.ResponseWriter, r *http.Request) {
			GetTwitterToken(w, r, s)
		},

		GetHomeTimeLine: func(w http.ResponseWriter, r *http.Request) {
			GetTimeLine(w, r, s, "home")
		},
		GetUserTimeLine: func(w http.ResponseWriter, r *http.Request) {
			GetTimeLine(w, r, s, "user")
		},
	}
}

func index(w http.ResponseWriter, r *http.Request, s *service.Service) {

	if !s.HasAuth(r) {
		session, err := s.CookieStore.Get(r, s.SessionName)
		if err != nil {
			fmt.Printf("can't get session: %s\n", err)
		}
		screenName := ""
		if session.Values["screen-name"] != nil {
			screenName = session.Values["screen-name"].(string)
		}

		fmt.Fprintf(w, "<BODY>%s<CENTER><a href='/request'>twitter login</a></CENTER></BODY>", screenName)
		return
	} else {
		//Logon, redirect to display time line
		timelineURL := fmt.Sprintf("http://%s/time", r.Host)
		http.Redirect(w, r, timelineURL, http.StatusTemporaryRedirect)
	}
}

// InitTwitterLogin - initializes the OAuth login process
func InitTwitterLogin(w http.ResponseWriter, r *http.Request, s *service.Service) {
	fmt.Println("Enter redirect to twitter")
	fmt.Println("Token URL=", CallbackURL)

	reqUrl, reqToken, err := s.GetAuthURL(CallbackURL)
	if err != nil {
		log.Printf("error getting auth url: %s\n", err)
	}

	fmt.Printf("token: %+v\n", reqToken)

	session, err := s.CookieStore.Get(r, s.SessionName)
	if err != nil {
		fmt.Printf("can't get session: %s\n", err)
	}
	session.Values["req-token"] = *reqToken
	err = session.Save(r, w)
	if err != nil {
		fmt.Printf("err storring into the session: %s\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Println("Leave redirect")
	http.Redirect(w, r, reqUrl, http.StatusTemporaryRedirect)

}

func GetTwitterToken(w http.ResponseWriter, r *http.Request, s *service.Service) {
	fmt.Println("Enter Get twitter token")
	values := r.URL.Query()

	fmt.Printf("values: %+v\n", values)

	verificationCode := values.Get("oauth_verifier")
	tokenKey := values.Get("oauth_token")

	session, err := s.CookieStore.Get(r, s.SessionName)
	if err != nil {
		fmt.Printf("err retrieving the session: %s\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	//reqToken := session.Values["req-token"].(oauth.RequestToken)
	reqToken, exists := s.OAuthTokens[tokenKey]
	if !exists || reqToken == nil {
		fmt.Printf("where's the token?!?!\n")
		fmt.Printf("err retrieving the req token: %s\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	accessToken, err := s.CompleteAuth(reqToken, verificationCode)
	if err != nil {
		fmt.Printf("err retrieving the access token: %s\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if accessToken != nil {
		user, err := s.VerifyCredentials(accessToken)
		if err != nil {
			fmt.Printf("error retrieving user info: %s\n", err)
		}

		session.Values["screen-name"] = user.ScreenName
		session.Values["acc-token"] = *accessToken
		err = session.Save(r, w)
		if err != nil {
			fmt.Printf("err storring accessToken into the session: %s\n", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	timelineURL := fmt.Sprintf("http://%s/time", r.Host)

	http.Redirect(w, r, timelineURL, http.StatusTemporaryRedirect)
}

func GetTimeLine(w http.ResponseWriter, r *http.Request, s *service.Service, which string) {
	aToken, err := s.GetAccessTokenFromSession(r)
	if err != nil || aToken == nil {
		fmt.Printf("err getting access token: %s\n", err)
		//http.Error(w, err.Error(), http.StatusInternalServerError)
		loginURL := fmt.Sprintf("http://%s/request", r.Host)
		http.Redirect(w, r, loginURL, http.StatusTemporaryRedirect)
		return
	}

	fmt.Printf("token: %+v\n\n\n\n\n", aToken)

	//var err error
	tweets := twitter.TimelineTweets{}
	switch which {
	case "user":
		tweets, err = s.UserTimeline(aToken, 20)
	case "home":
		tweets, err = s.QueryTimeLine(aToken, 20)
	}

	if err != nil {
		fmt.Printf("timeline err: %s\n", err)
	}

	w.Header().Add("Content-type", "text/plain")
	for _, t := range tweets {
		fmt.Fprintf(w, "%s // @%s\n", t.CreatedAt, t.User.ScreenName)
		fmt.Fprintf(w, "%+v\n", t.Entities)
		fmt.Fprintf(w, "%s\n", t.Text)
		fmt.Fprintln(w, "-------------")
	}
}
