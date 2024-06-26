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

	lis, err := net.Listen("tcp", ":9699")

	if err != nil {
		log.Fatal(err.Error())
	}

	go newServ.Serve(lis)

	go func() {
		if err := newServ.Serve(lis); err != nil {
			log.Fatal(err.Error())
		}
	}()

	cond := true

	log.Println("Kafka listening")

	for cond { // nolint:all

		ev := serv.Consumer_views.Poll(10)

		switch e := ev.(type) {

		case *kafka.Message:
			log.Println("new message for view")
			log.Println(ev)

			// extract data
			var data common.Reaction
			json.Unmarshal(e.Value, &data)
			// write to db
			ts, err := serv.Click.Begin()

			if err != nil {
				log.Println(err.Error())
				ts.Rollback()
				continue
			}

			query_doubler_check := "SELECT * FROM views WHERE user = ? AND post_id = ?"

			row := ts.QueryRow(query_doubler_check, data.User, data.PostId)

			var usr string
			var id uint64
			err = row.Scan(&usr, &id)

			if err == nil {
				log.Println("Duplicate view from ", data.User, "on post_id ", data.PostId, "with author ", data.Author)
				ts.Rollback()
				continue
			}
			log.Println(err.Error())
			_, err = ts.Exec("INSERT INTO views (user, post_id) VALUES (?, ?)", data.User, data.PostId)

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
		}

		ev = serv.Consumer_likes.Poll(10)

		switch e := ev.(type) {

		case *kafka.Message:
			log.Println("new message for likes")
			log.Println(ev)

			// extract data
			var data common.Reaction
			json.Unmarshal(e.Value, &data)
			// write to db
			ts, err := serv.Click.Begin()

			if err != nil {
				log.Fatal(err.Error())
				ts.Rollback()
				continue
			}

			query_doubler_check := "SELECT * FROM likes WHERE user = ? AND post_id = ?"

			row := ts.QueryRow(query_doubler_check, data.User, data.PostId)
			var usr string
			var id uint64
			var auth string
			err = row.Scan(&usr, &id, &auth)

			if err == nil {
				log.Println("Duplicate like from ", data.User, "on post_id ", data.PostId, "with author ", data.Author)
				log.Println("Removing like")

				_, err = ts.Exec("ALTER TABLE likes DELETE WHERE user = ? AND post_id = ?", data.User, data.PostId)

				if err != nil {
					log.Println("Failed to remove like")
					ts.Rollback()
					continue
				}

				ts.Commit()
				continue
			}

			_, err = ts.Exec("INSERT INTO likes (user, post_id, author) VALUES (?, ?, ?)", data.User, data.PostId, data.Author)

			if err != nil {
				log.Println("Failed to write view")
				log.Println(err.Error())
				ts.Rollback()
				continue
			}

			ts.Commit()

		case *kafka.Error:
			log.Fatal(e)
			// cond = false
			// cond = true

		default:
		}

	}
}
