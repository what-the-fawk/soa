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

	http.HandleFunc("/update", serv.Update)
	log.Println("Update")

	err := http.ListenAndServe(":3333", nil)

	if err != nil {
		log.Println(err.Error())
	}

	defer serv.Close()
}
