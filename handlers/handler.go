package handlers

import (
	"fmt"
	"log"
	"net/http"
	"twitwo/service"

	"github.com/mrjones/oauth"
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
	GetTimeLine      http.HandlerFunc
}

func NewHandlers(s *service.ServerClient) Handlers {
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

		GetTimeLine: func(w http.ResponseWriter, r *http.Request) {
			GetTimeLine(w, r, s)
		},
	}
}

func index(w http.ResponseWriter, r *http.Request, s *service.ServerClient) {

	if !s.HasAuth(r) {
		fmt.Fprintf(w, "<BODY><CENTER><A HREF='/request'><IMG SRC='https://g.twimg.com/dev/sites/default/files/images_documentation/sign-in-with-twitter-gray.png'></A></CENTER></BODY>")
		return
	} else {
		//Logon, redirect to display time line
		timelineURL := fmt.Sprintf("http://%s/time", r.Host)
		http.Redirect(w, r, timelineURL, http.StatusTemporaryRedirect)
	}
}

// InitTwitterLogin - initializes the OAuth login process
func InitTwitterLogin(w http.ResponseWriter, r *http.Request, s *service.ServerClient) {
	fmt.Println("Enter redirect to twitter")
	fmt.Println("Token URL=", CallbackURL)

	reqUrl, reqToken, err := s.GetAuthURL(CallbackURL)
	if err != nil {
		log.Printf("error getting auth url: %s\n", err)
	}

	session, _ := s.CookieStore.Get(r, s.SessionName)
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

func GetTwitterToken(w http.ResponseWriter, r *http.Request, s *service.ServerClient) {
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

	reqToken := session.Values["req-token"].(oauth.RequestToken)
	if reqToken.Token != tokenKey {
		fmt.Printf("Hmmm, got different things: %s - %s\n", reqToken.Token, tokenKey)
	}

	accessToken, err := s.CompleteAuth(&reqToken, verificationCode)
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

		fmt.Printf("username: @%s\n", user.ScreenName)
		fmt.Printf("user: %+v\n\n", user)

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

func GetTimeLine(w http.ResponseWriter, r *http.Request, s *service.ServerClient) {
	aToken, err := s.GetAccessTokenFromSession(r)
	if err != nil {
		fmt.Printf("err getting access token: %s\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	tweets, bits, _ := s.QueryTimeLine(aToken, 2)
	fmt.Println("TimeLine=", tweets)
	fmt.Printf("-- client: ---\n%+v\n-----\n", s)
	fmt.Printf("-- OAuthConsumer: ---\n%+v\n-----\n", *s.OAuthConsumer)
	// for k, _ := range twitterService.OAuthTokens {
	// 	fmt.Printf("\t%s:  %+v\n-----\n", k, *twitterService.OAuthTokens[k])
	// }

	for _, t := range tweets {
		fmt.Printf("* #%s -- @%s\n%s\n%s\n\n", t.ID, t.User.ScreenName, t.Entities, t.Text)
	}

	fmt.Fprintf(w, "The item is: %s", bits)
}
