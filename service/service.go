package service

import (
	"encoding/gob"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/sessions"
	"github.com/kkdai/twitter"
	"github.com/mrjones/oauth"
)

const (
	//Basic OAuth related URLs
	OAUTH_REQUES_TOKEN string = "https://api.twitter.com/oauth/request_token"
	OAUTH_AUTH_TOKEN   string = "https://api.twitter.com/oauth/authorize"
	OAUTH_ACCESS_TOKEN string = "https://api.twitter.com/oauth/access_token"

	SESSION_NAME string = "twidder"

	//List API URLs
	API_BASE           string = "https://api.twitter.com/1.1/"
	API_TIMELINE       string = API_BASE + "statuses/home_timeline.json"
	API_FOLLOWERS_IDS  string = API_BASE + "followers/ids.json"
	API_FOLLOWERS_LIST string = API_BASE + "followers/list.json"
	API_FOLLOWER_INFO  string = API_BASE + "users/show.json"
	API_ACCOUNT_INFO   string = API_BASE + "account/verify_credentials.json"
)

func init() {
	gob.Register(oauth.RequestToken{})
	gob.Register(oauth.AccessToken{})
}

func NewServerClient(consumerKey, consumerSecret string) *ServerClient {
	//newClient := NewClient(consumerKey, consumerKey)

	newServer := &ServerClient{
		OAuthConsumer: oauth.NewConsumer(
			consumerKey,
			consumerSecret,
			oauth.ServiceProvider{
				RequestTokenUrl:   OAUTH_REQUES_TOKEN,
				AuthorizeTokenUrl: OAUTH_AUTH_TOKEN,
				AccessTokenUrl:    OAUTH_ACCESS_TOKEN,
			},
		),
		OAuthTokens: make(map[string]*oauth.RequestToken),
		CookieStore: sessions.NewCookieStore([]byte(os.Getenv("SESSION_KEY"))),
		SessionName: SESSION_NAME,
	}

	//Enable debug info
	newServer.OAuthConsumer.Debug(false)

	return newServer
}

type ServerClient struct {
	//twitter.Client
	OAuthConsumer *oauth.Consumer
	OAuthTokens   map[string]*oauth.RequestToken
	CookieStore   *sessions.CookieStore
	SessionName   string
}

func (s *ServerClient) GetAuthURL(tokenUrl string) (string, *oauth.RequestToken, error) {
	token, requestUrl, err := s.OAuthConsumer.GetRequestTokenAndUrl(tokenUrl)
	if err != nil {
		log.Printf("error in GetRequestTokenAndUrl(): %s\n", err)
		return "", nil, errors.New("can't get the request token")
	}

	// Make sure to save the token, we'll need it for AuthorizeToken()
	s.OAuthTokens[token.Token] = token
	return requestUrl, token, nil
}

func (s *ServerClient) CompleteAuth(reqToken *oauth.RequestToken, verificationCode string) (*oauth.AccessToken, error) {
	accessToken, err := s.OAuthConsumer.AuthorizeToken(reqToken, verificationCode)
	if err != nil {
		log.Printf("error completing the auth: %s\n", err)
		return nil, err
	}

	fmt.Printf("ServerClient.CompleteAuth(): got accessToken: %s\n", accessToken)

	// s.HttpConn, err = s.OAuthConsumer.MakeHttpClient(accessToken)
	// if err != nil {
	// 	log.Println(err)
	// 	return err
	// }
	return accessToken, nil
}

func (s *ServerClient) GetAccessTokenFromSession(r *http.Request) (*oauth.AccessToken, error) {
	session, err := s.CookieStore.Get(r, s.SessionName)
	if err != nil {
		fmt.Printf("BuildClient(r): err retrieving the session object: %s\n", err)
		// http.Error(w, err.Error(), http.StatusInternalServerError)
		return nil, errors.New("can't build auth client")
	}

	token := session.Values["acc-token"]
	if token == nil {
		return nil, nil
	}
	aToken := token.(oauth.AccessToken)
	return &aToken, nil
}

func (s *ServerClient) BuildClient(accessToken *oauth.AccessToken) (*http.Client, error) {

	conn, err := s.OAuthConsumer.MakeHttpClient(accessToken)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	return conn, nil
}

//func (s *ServerClient) QueryTimeLinex(n int) (*http.Client, error) {

func (s *ServerClient) HasAuth(r *http.Request) bool {
	tk, err := s.GetAccessTokenFromSession(r)
	if err != nil || tk == nil {
		return false
	}
	return true
}

func (s *ServerClient) BasicQuery(token *oauth.AccessToken, queryString string) ([]byte, error) {
	c, err := s.BuildClient(token)
	if err != nil {
		fmt.Printf("error building the client: %s\n", err)
	}

	response, err := c.Get(queryString)
	if err != nil {
		log.Fatal(err)
	}
	defer response.Body.Close()

	bits, err := ioutil.ReadAll(response.Body)
	return bits, err
}

func (c *ServerClient) QueryTimeLine(token *oauth.AccessToken, count int) (twitter.TimelineTweets, []byte, error) {
	requestURL := fmt.Sprintf("%s?count=%d", API_TIMELINE, count)
	data, err := c.BasicQuery(token, requestURL)
	if err != nil {
		fmt.Printf("err: %s\n", err)
	}
	ret := twitter.TimelineTweets{}
	err = json.Unmarshal(data, &ret)
	return ret, data, err
}

func (c *ServerClient) VerifyCredentials(token *oauth.AccessToken) (twitter.UserDetail, error) {
	requestURL := fmt.Sprintf("%s?skip_status=true&include_email=true", API_ACCOUNT_INFO)
	data, err := c.BasicQuery(token, requestURL)
	if err != nil {
		fmt.Printf("err verifying credentials: %s\n", err)
	}
	user := twitter.UserDetail{}
	err = json.Unmarshal(data, &user)
	return user, err
}
