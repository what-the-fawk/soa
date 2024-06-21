package main

import (
	"log"
	"net/http"
	service "soa/main_service/include"
	"time"
)

func main() {

	time.Sleep(time.Second * 5) // wait for DB

	serv := service.CreateMainServiceHandler()
	log.Println("created")

	http.HandleFunc("/register", serv.Register)
	log.Println("Register")

	http.HandleFunc("/auth", serv.Auth)
	log.Println("Auth")

	http.HandleFunc("/users", serv.Update)
	log.Println("Update")

	http.HandleFunc("/create", serv.CreatePost)
	log.Println("Create")

	http.HandleFunc("/update", serv.UpdatePost)
	log.Println("Update post")

	http.HandleFunc("/getPost", serv.GetPost)
	log.Println("Get post")

	http.HandleFunc("/delete", serv.DeletePost)
	log.Println("Delete")

	http.HandleFunc("/list", serv.GetPostList)
	log.Println("Get post list")

	http.HandleFunc("/like", serv.SendLike)
	log.Println("Send like")

	http.HandleFunc("/view", serv.SendView)
	log.Println("Send view")

	http.HandleFunc("/topP", serv.TopPosts)
	log.Println("Top posts")

	http.HandleFunc("/total", serv.TotalActivity)
	log.Println("Total Activity")

	http.HandleFunc("/topU", serv.TopUsers)
	log.Println("Top users")

	err := http.ListenAndServe(":3333", nil)

	if err != nil {
		log.Println(err.Error())
	}

	defer serv.Close()
}
