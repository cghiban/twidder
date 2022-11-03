package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"

	"twidel/handlers"
	"twidel/service"
)

func init() {

}

func main() {

	consumerKey := os.Getenv("CONSUMER_KEY")
	consumerSecret := os.Getenv("CONSUMER_SECRET")

	if consumerKey == "" && consumerSecret == "" {
		fmt.Println("missing CONSUMER_KEY and CONSUMER_SECRET.")
		return
	}

	var port *int = flag.Int("port", 9090, "Port to listen on.")
	flag.Parse()

	svc := service.NewService(consumerKey, consumerSecret)

	handlers := handlers.NewHandlers(svc)

	http.HandleFunc("/maketoken", handlers.GetTwitterToken)
	http.HandleFunc("/request", handlers.InitTwitterLogin)
	http.HandleFunc("/time", handlers.GetHomeTimeLine)
	http.HandleFunc("/my", handlers.GetUserTimeLine)
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
