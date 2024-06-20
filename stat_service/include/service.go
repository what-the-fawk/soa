package statservice

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"soa/stat_service/stats_service/pkg/pb"
	"time"

	"database/sql"

	_ "github.com/ClickHouse/clickhouse-go"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

type StatServiceHandler struct {
	pb.UnimplementedStatServiceServer
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

	_, err = db.Exec("CREATE TABLE IF NOT EXISTS views (user String, post_id UInt64) ENGINE = MergeTree() ORDER BY post_id") // TODO: change engines

	if err != nil {
		log.Println("Failed to create views")
		log.Fatal(err.Error())
	}

	// create likes

	_, err = db.Exec("CREATE TABLE IF NOT EXISTS likes (user String, post_id UInt64, author String) ENGINE = MergeTree() ORDER BY post_id")

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

func (s *StatServiceHandler) Top(ctx context.Context, info *pb.TopInfo) (*pb.Posts, error) {

	// go to db
	log.Println("Top")

	ts, err := s.Click.Begin()

	if err != nil {
		log.Println("Clickhouse Begin issues")
		return nil, err
	}

	var table string

	if info.IsLike != 0 {
		table = "likes"
	} else {
		table = "views"
	}

	query := "SELECT post_id, COUNT(post_id) AS `value_occurence` FROM " + table + " GROUP BY post_id ORDER BY `value_occurence` DESC LIMIT 5"

	rows, err := ts.Query(query)

	if err != nil {
		return nil, err
	}

	ans := &pb.Posts{}

	for rows.Next() {
		var post pb.PostInfo

		post.AuthorLogin = "unknown"

		err = rows.Scan(&post.Id, &post.ActivityNum)

		if err != nil {
			return nil, err
		}

		ans.Ids = append(ans.Ids, &post)
	}

	ts.Commit()

	return ans, nil
}

func (s *StatServiceHandler) Rating(ctx context.Context, arg *emptypb.Empty) (*pb.Users, error) {

	// go to db
	log.Println("Rating")

	ts, err := s.Click.Begin()

	if err != nil {
		log.Println("Clickhouse Begin issues")
		return nil, err
	}

	query := "SELECT author, COUNT(author) AS `value_occurence` FROM likes GROUP BY author ORDER BY `value_occurence` DESC LIMIT 3"

	rows, err := ts.Query(query)

	if err != nil {
		return nil, err
	}

	ans := &pb.Users{}

	for rows.Next() {
		var info pb.UserLike

		err = rows.Scan(&info.Login, &info.Likes)

		log.Println("User: ", info.Likes, info.Login)

		if err != nil {
			return nil, err
		}

		ans.Users = append(ans.Users, &info)
	}

	ts.Commit()

	return ans, nil
}

func (s *StatServiceHandler) Total(ctx context.Context, id *pb.PostID) (*pb.LikesViews, error) {

	// go to db
	log.Println("Total")
	log.Println(id.Id)

	ts, err := s.Click.Begin()
	ans := &pb.LikesViews{}

	if err != nil {
		log.Println("Clickhouse Begin issues")
		return nil, err
	}

	query_likes := "SELECT COUNT(DISTINCT user) as count FROM likes WHERE post_id = ?"

	row := ts.QueryRow(query_likes, int64(id.Id))

	if row.Err() != nil {
		log.Println(row.Err().Error())
		return nil, row.Err()
	}

	log.Println("LikesViews")

	err = row.Scan(&ans.Likes)
	if err != nil {
		log.Println(err.Error())
		return nil, err
	}

	query_views := "SELECT COUNT(DISTINCT user) as count FROM views WHERE post_id = ?"

	row = ts.QueryRow(query_views, int64(id.Id))

	if row.Err() != nil {
		log.Println(row.Err().Error())
		return nil, row.Err()
	}

	err = row.Scan(&ans.Views)
	if err != nil {
		log.Println(err.Error())
		return nil, err
	}

	return ans, nil
}
