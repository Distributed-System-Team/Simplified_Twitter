package twitter

import (
	"Simplified_Twitter/src/auth/cookie"
	"fmt"
	// pb "google.golang.org/grpc/examples/Simplified_Twitter/src/twitter_web/TwitterPage"
	"html/template"
	"net/http"
	"Simplified_Twitter/src/rpc/client"
	"Simplified_Twitter/src/storage"
	"time"
)

func Twitter(w http.ResponseWriter, r *http.Request) {
	uName := cookie.GetUserName(r)
	if uName != "" {
		fmt.Println("----------------> Test rpc Start")
		curUser := client.RpcGetUser(uName)
		fmt.Println("----------------> Test rpc End")
		switch r.Method {
		case "GET":
			t, err := template.ParseFiles("./src/template/post.html")
			if err != nil {
				fmt.Fprintf(w, "Error : %v\n", err)
				return
			}
			t.Execute(w, curUser)
		case "POST":
			r.ParseForm()
			var curTwit = storage.TwitPosts{}
			curTwit.Contents = r.Form.Get("contents")
			// // If the post contents are empty not post
			if curTwit.Contents != "" {
				curTwit.Date = time.Now().Unix()
				curTwit.User = uName
				curUser.Posts = append(curUser.Posts, curTwit) // TODO

				client.RpcUpdateUser(uName, curUser)
				fmt.Println("Posts", curUser.Posts)
			}
			http.Redirect(w, r, "/profile", 302)
		}
	} else {
		http.Redirect(w, r, "/", 302)
	}
}
