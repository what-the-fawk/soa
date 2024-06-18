package main

import (
	"encoding/json"
	"log"
	"net"
	"net/http"
	"soa/common"
	statservice "soa/stat_service/include"
	"soa/stat_service/stats_service/pkg/pb"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"google.golang.org/grpc"
)

func main() {

	time.Sleep(time.Second * 5) // wait for DB

	serv := statservice.CreateNewStatService()
	http.HandleFunc("/", serv.OK)

	newServ := grpc.NewServer()
	pb.RegisterStatServiceServer(newServ, serv)

	go http.ListenAndServe(":4857", nil)
	log.Println("Stat service")

	lis, err := net.Listen("tcp", ":2629")

	if err != nil {
		log.Fatal(err.Error())
	}

	go newServ.Serve(lis)

	cond := true

	for cond { // nolint:all

		ev := serv.Consumer_views.Poll(1000)

		switch e := ev.(type) {

		case *kafka.Message:
			log.Println("new message for view")
			log.Println(ev)

			// extract data
			var data common.ReactionInfo
			json.Unmarshal(e.Value, &data)
			// write to db
			ts, err := serv.Click.Begin()

			if err != nil {
				log.Fatal(err.Error())
			}

			_, err = ts.Exec("INSERT INTO views (user, post_id) VALUES (?, ?)", data.Author, data.PostId)

			if err != nil {
				log.Println("Failed to write view")
				log.Println(err.Error())
			}

			ts.Commit()

		case *kafka.Error:
			log.Println(e)
			// cond = false
			// cond = true

		default:
			log.Println("no message for views")
		}

		ev = serv.Consumer_likes.Poll(1000)

		switch e := ev.(type) {

		case *kafka.Message:
			log.Println("new message for likes")
			log.Println(ev)

			// extract data
			var data common.ReactionInfo
			json.Unmarshal(e.Value, &data)
			// write to db
			ts, err := serv.Click.Begin()

			if err != nil {
				log.Fatal(err.Error())
			}

			_, err = ts.Exec("INSERT INTO likes (user, post_id) VALUES (?, ?)", data.Author, data.PostId)

			if err != nil {
				log.Println("Failed to write view")
				log.Println(err.Error())
			}

			ts.Commit()

		case *kafka.Error:
			log.Fatal(e)
			// cond = false
			// cond = true

		default:
			log.Println("no message for likes")
		}

	}
}
