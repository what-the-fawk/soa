package main

import (
	"encoding/json"
	"log"
	"net/http"
	"soa/common"
	statservice "soa/stat_service/include"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

func main() {

	time.Sleep(time.Second * 5) // wait for DB

	serv := statservice.CreateNewStatService()
	http.HandleFunc("/", serv.OK)

	go http.ListenAndServe(":4857", nil)
	log.Println("Stat service")

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

			_, err = ts.Exec("INSERT INTO views (author_id, post_id) VALUES (?, ?)", data.AuthorId, data.PostId)

			ts.Commit()

			if err != nil {
				log.Println("Failed to write view")
				log.Println(err.Error())
			}

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

			_, err = ts.Exec("INSERT INTO likes (author_id, post_id) VALUES (?, ?)", data.AuthorId, data.PostId)

			ts.Commit()

			if err != nil {
				log.Println("Failed to write view")
				log.Println(err.Error())
			}

		case *kafka.Error:
			log.Fatal(e)
			// cond = false
			// cond = true

		default:
			log.Println("no message for likes")
		}

	}
}
