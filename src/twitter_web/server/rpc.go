package main

import (
	"context"
	"fmt"
	"google.golang.org/grpc"
	pb "google.golang.org/grpc/examples/Simplified_Twitter/src/twitter_web/TwitterPage"
	"google.golang.org/grpc/reflection"
	"sort"
	"sync"
	// "io"
	"log"
	"net"
)

const (
	port = ":9091"
)

// type server struct{}

// var WebDB DB

type User struct {
	UserName  string
	passWord  string
	Posts     Twitlist
	Following []string
}

// Define this type for sort
type Twitlist []TwitPosts

// Using time to define post order
// Username to who do the posts

type TwitPosts struct {
	Contents string
	Date     int64
	User     string
}

type TwitterPage struct {
	UserName   string
	UnFollowed []string
	Following  []string
	Posts      []string
	// Posts Twitlist
}

type DB struct {
	mu        sync.Mutex
	UsersInfo map[string]User
}

// Sort Function needed these three Function
func (I Twitlist) Len() int {
	return len(I)
}
func (I Twitlist) Less(i, j int) bool {
	return I[i].Date < I[j].Date
}
func (I Twitlist) Swap(i, j int) {
	I[i], I[j] = I[j], I[i]
}

func (db *DB) GetUser(ctx context.Context, in *pb.GetUserRequest) (*pb.GetUserReply, error) {
	db.mu.Lock()
	defer db.mu.Unlock()
	var uName string = in.Uname
	var tmp User = db.UsersInfo[uName]
	var posts []*pb.TwitPosts
	for _, post := range tmp.Posts {
		var tmpPost = &pb.TwitPosts{Contents: post.Contents, Date: post.Date, User: post.User}
		posts = append(posts, tmpPost)
	}
	var user = &pb.User{UserName: tmp.UserName, PassWord: tmp.passWord, Posts: posts, Following: tmp.Following}
	// log.Printf("------> server user", user)
	return &pb.GetUserReply{Userinfo: user}, nil
}

func (db *DB) AddUser(ctx context.Context, in *pb.AddUserRequest) (*pb.BoolReply, error) {
	db.mu.Lock()
	defer db.mu.Unlock()
	uName := in.Username
	pWord1 := in.Password1
	pWord2 := in.Password2
	if pWord1 != pWord2 {
		return &pb.BoolReply{T: false}, nil
	}
	if uName == "" || pWord1 == "" {
		return &pb.BoolReply{T: false}, nil
	}
	curUser := User{uName, pWord1, Twitlist{}, []string{uName}}
	if _, ok := db.UsersInfo[uName]; ok {
		return &pb.BoolReply{T: false}, nil
	}
	// Use uName as key put curUser inside
	db.UsersInfo[uName] = curUser
	return &pb.BoolReply{T: true}, nil
}

func (db *DB) UpdateUser(ctx context.Context, in *pb.UpdateUserRequest) (*pb.BoolReply, error) {
	db.mu.Lock()
	defer db.mu.Unlock()
	uName := in.Username
	var posts Twitlist
	for _, post := range in.Usr.Posts {
		var tmpPost = TwitPosts{Contents: post.Contents, Date: post.Date, User: post.User}
		posts = append(posts, tmpPost)
	}
	var usr = User{UserName: in.Usr.UserName, passWord: in.Usr.PassWord, Posts: posts, Following: in.Usr.Following}
	if _, ok := db.UsersInfo[uName]; ok != true {
		return &pb.BoolReply{T: false}, nil
	}
	db.UsersInfo[uName] = usr
	return &pb.BoolReply{T: true}, nil
}

func (db *DB) HasUser(ctx context.Context, in *pb.HasUserRequest) (*pb.BoolReply, error) {
	db.mu.Lock()
	defer db.mu.Unlock()
	uName := in.Username
	pWord := in.Password
	if uName == "" || pWord == "" {
		return &pb.BoolReply{T: false}, nil
	}
	// Check Whether User in usersInfo
	user, exist := db.UsersInfo[uName]
	if exist && user.passWord == pWord {
		return &pb.BoolReply{T: true}, nil
	}
	return &pb.BoolReply{T: false}, nil
}

func Contains(a []string, x string) bool {
	for _, n := range a {
		if x == n {
			return true
		}
	}
	return false
}

func Deletes(a []string, x string) []string {
	var ret []string
	for _, n := range a {
		if x != n {
			ret = append(ret, n)
		}
	}
	return ret
}

func (db *DB) FollowUser(ctx context.Context, in *pb.FollowUserRequest) (*pb.BoolReply, error) {
	db.mu.Lock()
	defer db.mu.Unlock()
	uName := in.Username
	otherName := in.Othername
	if user, ok := db.UsersInfo[uName]; ok {
		if Contains(user.Following, otherName) {
			return &pb.BoolReply{T: false}, nil
		}
		user.Following = append(user.Following, otherName)
		db.UsersInfo[uName] = user
		return &pb.BoolReply{T: true}, nil
	}
	return &pb.BoolReply{T: false}, nil
}

func (db *DB) UnFollowUser(ctx context.Context, in *pb.FollowUserRequest) (*pb.BoolReply, error) {
	db.mu.Lock()
	defer db.mu.Unlock()
	uName := in.Username
	otherName := in.Othername
	if user, ok := db.UsersInfo[uName]; ok {
		if !Contains(user.Following, otherName) {
			return &pb.BoolReply{T: false}, nil
		}
		user.Following = Deletes(user.Following, otherName)
		db.UsersInfo[uName] = user
		return &pb.BoolReply{T: true}, nil
	}
	return &pb.BoolReply{T: false}, nil
}

// // Get Rid of the arrtribute of time
// // Just leave username + contents

func GetContents(arr Twitlist) []string {
	var ret []string
	for _, twit := range arr {
		tmp := twit.User + ": " + twit.Contents
		ret = append(ret, tmp)
	}
	return ret
}

func (db *DB) GetTwitterPage(ctx context.Context, in *pb.GetTwitterPageRequest) (*pb.GetTwitterPageReply, error) {
	db.mu.Lock()
	defer db.mu.Unlock()
	uName := in.Username
	user, _ := db.UsersInfo[uName]
	log.Printf("-------> TwitterPage Userinfo ", user)
	UserName := user.UserName
	Following := user.Following
	log.Printf("..............", Following)
	var UnFollowed []string
	var Posts Twitlist
	// Get all Posts information
	for name, userInfo := range db.UsersInfo {
		if Contains(Following, name) {
			for _, post := range userInfo.Posts {
				Posts = append(Posts, post)
			}
		} else {
			UnFollowed = append(UnFollowed, name)
		}
	}
	fmt.Println(Posts)
	sort.Sort(Posts)
	newPosts := GetContents(Posts)
	// Remove the user itself from following list (just not shown in screen but in memory)
	Following = Deletes(Following, uName)
	log.Printf("------> TwitterPage Username %s", UserName)
	log.Printf("------> TwitterPage Following %s", Following)
	log.Printf("------> TwitterPage UnFollowed %s", UnFollowed)
	log.Printf("------> TwitterPage Posts %s", newPosts)
	var twit = &pb.TwitterPage{Username: UserName, UnFollowed: UnFollowed, Following: Following, Posts: newPosts}
	return &pb.GetTwitterPageReply{Twit: twit}, nil

}

func main() {
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	WebDB := &DB{}
	WebDB.UsersInfo = make(map[string]User)
	pb.RegisterWebServer(s, WebDB)
	// Register reflection service on gRPC server.
	reflection.Register(s)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
