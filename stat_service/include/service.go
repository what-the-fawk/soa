package statservice

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"database/sql"

	_ "github.com/ClickHouse/clickhouse-go"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

type StatServiceHandler struct {
	// db
	Consumer_views *kafka.Consumer
	Consumer_likes *kafka.Consumer
	Click          *sql.DB
}

func CreateNewStatService() *StatServiceHandler {

	time.Sleep(20 * time.Second)

	c, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers": "kafka:29092",
		"group.id":          "statistics",
	})

	if err != nil {
		log.Fatal(err.Error())
	}

	err = c.Subscribe("topic-view", nil)
	if err != nil {
		log.Fatal(err.Error())
	}

	log.Println("subscribed to view")

	cons, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers": "kafka:29092",
		"group.id":          "statistics",
	})
	if err != nil {
		log.Fatal(err.Error())
	}

	err = cons.Subscribe("topic-like", nil)
	if err != nil {
		log.Fatal(err.Error())
	}

	log.Println("subscribed to like")

	clik_host := "stat_db"
	click_name := os.Getenv("STAT_DB_USERNAME")
	click_pwd := os.Getenv("STAT_DB_PASSWORD")

	connect_str := fmt.Sprintf("http://%s:9000?username=%s&password=%s",
		clik_host, click_name, click_pwd)

	db, err := sql.Open("clickhouse", connect_str)

	if err != nil {
		log.Fatal(err.Error())
	}

	// create views

	_, err = db.Exec("CREATE TABLE IF NOT EXISTS views (author_id UInt64, post_id UInt64) ENGINE = Memory")

	if err != nil {
		log.Println("Failed to create views")
		log.Fatal(err.Error())
	}

	// create likes

	_, err = db.Exec("CREATE TABLE IF NOT EXISTS likes (author_id UInt64, post_id UInt64) ENGINE = Memory")

	if err != nil {
		log.Println("Failed to create likes")
		log.Fatal(err.Error())
	}

	return &StatServiceHandler{
		Consumer_views: c,
		Consumer_likes: cons,
		Click:          db,
	}
}

func (s *StatServiceHandler) OK(w http.ResponseWriter, req *http.Request) {

	if req.Method != http.MethodGet {
		http.Error(w, "Only get is allowed", http.StatusBadRequest)
		return
	}
	log.Println("GET OK")
	// returns OK
}
