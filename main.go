package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"

	"twitwo/handlers"
	"twitwo/service"
)

var ConsumerKey string
var ConsumerSecret string
var twitterService *service.ServerClient

func init() {
	ConsumerKey = os.Getenv("CONSUMER_KEY")
	ConsumerSecret = os.Getenv("CONSUMER_SECRET")
}

func main() {

	if ConsumerKey == "" && ConsumerSecret == "" {
		ConsumerKey = "UffiIkCapOH8vsuYcTJD5Kvzt4Z0tHvnm1NUH4WWqKKP6PbrNj"
		ConsumerSecret = "9ltl9Qs8AVx1hwab44xiHi1ut"
		//fmt.Println("Please setup ConsumerKey and ConsumerSecret.")
		//return
	}

	var port *int = flag.Int("port", 9090, "Port to listen on.")
	flag.Parse()

	fmt.Println("[app] Init server key=", ConsumerKey, " secret=", ConsumerSecret)
	twitterService = service.NewServerClient(ConsumerKey, ConsumerSecret)

	handlers := handlers.NewHandlers(twitterService)

	http.HandleFunc("/maketoken", handlers.GetTwitterToken)
	http.HandleFunc("/request", handlers.InitTwitterLogin)
	http.HandleFunc("/time", handlers.GetTimeLine)
	// http.HandleFunc("/follow", GetFollower)
	// http.HandleFunc("/followids", GetFollowerIDs)

	//http.HandleFunc("/user", GetUserDetail)
	http.HandleFunc("/", handlers.Index)

	u := fmt.Sprintf(":%d", *port)
	fmt.Printf("Listening on '%s'\n", u)
	http.ListenAndServe(u, nil)
}

/*
func GetFollower(w http.ResponseWriter, r *http.Request) {
	followers, bits, _ := twitterClient.QueryFollower(10)
	fmt.Println("Followers=", followers)
	fmt.Fprintf(w, "The item is: "+string(bits))
}

func GetFollowerIDs(w http.ResponseWriter, r *http.Request) {
	followers, bits, _ := twitterClient.QueryFollowerIDs(10)
	fmt.Println("Follower IDs=", followers)
	fmt.Fprintf(w, "The item is: "+string(bits))
}
func GetUserDetail(w http.ResponseWriter, r *http.Request) {
	followers, bits, _ := twitterClient.QueryFollowerById(2244994945)
	fmt.Println("Follower Detail of =", followers)
	fmt.Fprintf(w, "The item is: "+string(bits))
}
*/
