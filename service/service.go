package service

import (
	"encoding/base64"
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
	API_USER_TIMELINE  string = API_BASE + "statuses/user_timeline.json"
	API_FOLLOWERS_IDS  string = API_BASE + "followers/ids.json"
	API_FOLLOWERS_LIST string = API_BASE + "followers/list.json"
	API_FOLLOWER_INFO  string = API_BASE + "users/show.json"
	API_ACCOUNT_INFO   string = API_BASE + "account/verify_credentials.json"
)

func init() {

	gob.Register(oauth.RequestToken{})
	gob.Register(oauth.AccessToken{})
}

func NewService(consumerKey, consumerSecret string) *Service {

	var err error
	var authKey, encKey []byte
	authKeyStr := os.Getenv("SESSION_KEY")
	if authKeyStr == "" {
		log.Fatal("SESSION_KEY not set")
	}

	encKeyStr := os.Getenv("SESSION_ENC_KEY")
	if encKeyStr == "" {
		encKey = nil
	} else {
		encKey, err = base64.StdEncoding.DecodeString(authKeyStr)
		if err != nil {
			log.Fatalf("some error occured during base64 decoding SESSION_KEY. Error %s\n", err)
		}
	}

	authKey, err = base64.StdEncoding.DecodeString(authKeyStr)
	if err != nil {
		log.Fatalf("some error occured during base64 decoding SESSION_ENC_KEY. Error %s\n", err)
	}

	cStore := sessions.NewCookieStore(authKey, encKey)
	cStore.Options.HttpOnly = true
	//cStore.Options.Secure = true
	//cStore.Options.SameSite = http.SameSiteStrictMode

	newServer := &Service{
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
		CookieStore: cStore,
		SessionName: SESSION_NAME,
	}

	//Enable debug info
	newServer.OAuthConsumer.Debug(false)

	return newServer
}

type Service struct {
	_             struct{}
	OAuthConsumer *oauth.Consumer
	OAuthTokens   map[string]*oauth.RequestToken
	CookieStore   *sessions.CookieStore
	SessionName   string
}

func (s *Service) GetAuthURL(tokenUrl string) (string, *oauth.RequestToken, error) {
	token, requestUrl, err := s.OAuthConsumer.GetRequestTokenAndUrl(tokenUrl)
	if err != nil {
		log.Printf("error in GetRequestTokenAndUrl(): %s\n", err)
		return "", nil, errors.New("can't get the request token")
	}

	// Make sure to save the token, we'll need it for AuthorizeToken()
	s.OAuthTokens[token.Token] = token
	return requestUrl, token, nil
}

func (s *Service) CompleteAuth(reqToken *oauth.RequestToken, verificationCode string) (*oauth.AccessToken, error) {
	accessToken, err := s.OAuthConsumer.AuthorizeToken(reqToken, verificationCode)
	if err != nil {
		log.Printf("error completing the auth: %s\n", err)
		return nil, err
	}

	fmt.Printf("Service.CompleteAuth(): got accessToken: %s\n", accessToken)

	return accessToken, nil
}

func (s *Service) GetAccessTokenFromSession(r *http.Request) (*oauth.AccessToken, error) {
	session, err := s.CookieStore.Get(r, s.SessionName)
	if err != nil {
		fmt.Printf("BuildClient(r): err retrieving the session object: %s\n", err)
		// http.Error(w, err.Error(), http.StatusInternalServerError)
		return nil, errors.New("can't build auth client")
	}

	fmt.Println("--------------------------")
	for _, x := range session.Values {
		fmt.Printf(" ** %s:\t\n", x)
	}
	fmt.Println("--------------------------")

	token := session.Values["acc-token"]
	if token == nil {
		return nil, nil
	}
	aToken := token.(oauth.AccessToken)
	return &aToken, nil
}

func (s *Service) BuildClient(accessToken *oauth.AccessToken) (*http.Client, error) {

	conn, err := s.OAuthConsumer.MakeHttpClient(accessToken)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	return conn, nil
}

func (s *Service) HasAuth(r *http.Request) bool {
	tk, err := s.GetAccessTokenFromSession(r)
	if err != nil || tk == nil {
		return false
	}
	return true
}

func (s *Service) BasicQuery(token *oauth.AccessToken, uri string) ([]byte, error) {
	c, err := s.BuildClient(token)
	if err != nil {
		fmt.Printf("error building the client: %s\n", err)
	}

	fmt.Printf("++ gonna retrieve %s\n", uri)
	resp, err := c.Get(uri)
	c.CloseIdleConnections()
	if err != nil {
		//log.Fatal(err)
		return nil, err
	}
	defer resp.Body.Close()
	bits, err := ioutil.ReadAll(resp.Body)

	if resp.StatusCode == http.StatusUnauthorized {
		fmt.Printf("%d: %s", resp.StatusCode, bits)
		return nil, errors.New("unauthorized")
	}
	if resp.StatusCode/100 != 2 {
		fmt.Printf("ERR: RESP CODE: %d\n", resp.StatusCode)
		return nil, fmt.Errorf("error with code %d", resp.StatusCode)
	}

	return bits, err
}

func (c *Service) QueryTimeLine(token *oauth.AccessToken, count int) (twitter.TimelineTweets, error) {
	requestURL := fmt.Sprintf("%s?count=%d", API_TIMELINE, count)
	data, err := c.BasicQuery(token, requestURL)
	if err != nil {
		fmt.Printf("err: %s\n", err)
	}
	ret := twitter.TimelineTweets{}
	err = json.Unmarshal(data, &ret)
	return ret, err
}

func (c *Service) VerifyCredentials(token *oauth.AccessToken) (twitter.UserDetail, error) {
	requestURL := fmt.Sprintf("%s?skip_status=true&include_email=true", API_ACCOUNT_INFO)
	data, err := c.BasicQuery(token, requestURL)
	if err != nil {
		fmt.Printf("err verifying credentials: %s\n", err)
	}
	user := twitter.UserDetail{}
	err = json.Unmarshal(data, &user)
	return user, err
}

func (c *Service) UserTimeline(token *oauth.AccessToken, count int) (twitter.TimelineTweets, error) {
	requestURL := fmt.Sprintf("%s?count=%d", API_USER_TIMELINE, count)
	data, err := c.BasicQuery(token, requestURL)
	if err != nil {
		fmt.Printf("err UserTimeline(): %s\n", err)
		return nil, err
	}
	ret := twitter.TimelineTweets{}
	err = json.Unmarshal(data, &ret)
	return ret, err
}
