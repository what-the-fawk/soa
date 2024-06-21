package service

import (
	"context"
	"crypto/rsa"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"soa/common"
	"soa/post_service/posts_service/pkg/pb"
	"time"

	rpc_stats "soa/stat_service/stats_service/pkg/pb"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/golang-jwt/jwt/v5"
	_ "github.com/lib/pq"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/emptypb"
)

type MainServiceHandler struct {
	Db         *sql.DB
	JwtPrivate *rsa.PrivateKey
	JwtPublic  *rsa.PublicKey
	client     pb.PostServiceClient
	stats      rpc_stats.StatServiceClient
	producer   *kafka.Producer
	admin      *kafka.AdminClient
}

const dbname = "postgres"
const connectionStringPattern string = "host=%s port=%d user=%s password=%s dbname=%s sslmode=%s"

func CreateMainServiceHandler() *MainServiceHandler {

	host, port, user, password, sslmode, err := common.GetPostgresParams()

	if err != nil {
		log.Fatal(err.Error())
	}

	connectStr := fmt.Sprintf(connectionStringPattern,
		host, port, user, password, dbname, sslmode)

	log.Println("Connecting...")

	db, err := sql.Open(dbname, connectStr)

	for i := 0; i < 10; i++ {

		if err == nil {
			break
		}

		log.Println("Connecting...")

		time.Sleep(time.Second * 2)

		db, err = sql.Open(dbname, connectStr)
	}

	if err != nil {
		log.Fatal(err.Error())
	}

	err = db.Ping()

	if err != nil {
		log.Fatal(err.Error())
	}

	const query = "" +
		"CREATE TABLE IF NOT EXISTS Users " +
		"(" +
		"login VARCHAR (50) UNIQUE NOT NULL, " +
		"password TEXT NOT NULL, " +
		"first_name VARCHAR (40), " +
		"second_name VARCHAR (40), " +
		"date_of_birth VARCHAR (50), " +
		"email VARCHAR (40), " +
		"phone_number VARCHAR (40)" +
		")"

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	_, err = db.ExecContext(ctx, query)

	if err != nil {
		log.Fatal(err.Error())
	}

	pub, pri, err := common.GetRSAKeys()

	if err != nil {
		log.Println("Rsa keys error")
		log.Fatal(err.Error())
	}

	time.Sleep(20 * time.Second)

	grpcServerAddr, ok := os.LookupEnv("GRPC_SERVER")
	if !ok {
		log.Fatal("GRPC_SERVER not set")
	}
	conn, err := grpc.Dial(grpcServerAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect to gRPC server: %v", err.Error())
	}
	grpcClient := pb.NewPostServiceClient(conn)

	grpcPostsAddr := "stat_service:9699"
	conn, err = grpc.Dial(grpcPostsAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))

	if err != nil {
		log.Fatalf("Failed to connect to gRPC server stats: %v", err.Error())
	}

	grpcPosts := rpc_stats.NewStatServiceClient(conn)

	p, err := kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers": "kafka:29092",
		"acks":              "all"})

	if err != nil {
		log.Fatalf("Failed to create producer: %s\n", err)
	}

	a, err := kafka.NewAdminClientFromProducer(p)

	if err != nil {
		log.Fatalf("Failed to create admin client: %s\n", err.Error())
	}

	var topics []kafka.TopicSpecification

	topics = append(topics, kafka.TopicSpecification{
		Topic:             "topic-view",
		NumPartitions:     1,
		ReplicationFactor: 1,
		ReplicaAssignment: nil,
		Config:            nil,
	})

	topics = append(topics, kafka.TopicSpecification{
		Topic:             "topic-like",
		NumPartitions:     1,
		ReplicationFactor: 1,
		ReplicaAssignment: nil,
		Config:            nil,
	})

	_, err = a.CreateTopics(context.Background(), topics)

	if err != nil {
		log.Fatalf("Failed to create topics: %s\n", err.Error())
	}

	return &MainServiceHandler{
		Db:         db,
		JwtPublic:  pub,
		JwtPrivate: pri,
		client:     grpcClient,
		stats:      grpcPosts,
		producer:   p,
		admin:      a,
	}
}

func (s *MainServiceHandler) Close() {
	s.Db.Close()
}

func (s *MainServiceHandler) Register(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		log.Println("Wrong method in Register")
		http.Error(w, "Registration is allowed only with POST method", http.StatusBadRequest)
		return
	}

	info, status, err := common.GetJsonStruct[common.AuthInfo](req)

	if err != nil {
		log.Println("Json unmarshall error")
		http.Error(w, err.Error(), status)
		return
	}

	const query = "" +
		"INSERT INTO Users " +
		"(login, password) " +
		"VALUES ($1, $2)"

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	defer cancel()

	hasher := sha256.New()
	passwordHash := hex.EncodeToString(hasher.Sum([]byte(info.Password)))

	_, err = s.Db.ExecContext(ctx, query, info.Login, passwordHash)

	if err != nil {
		log.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (s *MainServiceHandler) Auth(w http.ResponseWriter, req *http.Request) {

	if req.Method != http.MethodPost {
		log.Println("Wrong method in Auth")
		http.Error(w, "Authentication is allowed only with GET method", http.StatusBadRequest)
		return
	}

	info, status, err := common.GetJsonStruct[common.AuthInfo](req)

	if err != nil {
		log.Println(err.Error())
		http.Error(w, err.Error(), status)
		return
	}

	const query = "SELECT login, password from Users WHERE login=$1"

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	defer cancel()

	row := s.Db.QueryRowContext(ctx, query, info.Login)

	userQueryInfo := common.AuthInfo{}

	err = row.Scan(&userQueryInfo.Login, &userQueryInfo.Password)

	if err != nil {
		log.Println("Row scan error", err.Error())
		http.Error(w, "Incorrect login", http.StatusNotFound)
		return
	}

	hasher := sha256.New()
	passwordHash := hex.EncodeToString(hasher.Sum([]byte(info.Password)))

	if string(passwordHash[:]) != userQueryInfo.Password {
		log.Println("Incorrect password")
		http.Error(w, "Incorrect password", http.StatusNotFound)
		return
	}

	// token gen
	claims := jwt.MapClaims{
		"iss": info.Login,
		"exp": time.Duration(time.Now().Second()) + time.Duration(1800*time.Second),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)

	tokenStr, err := token.SignedString(s.JwtPrivate)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:  "jwt",
		Value: tokenStr,
	})
}

// check token validity, returns jwt issuer and error
func (s *MainServiceHandler) CheckToken(req *http.Request) (string, error) {

	// nolint:all
	cookie, err := req.Cookie("jwt")

	if err != nil {
		log.Println("No jwt?")
		return "", errors.New("No jwt?")
	}

	tokenStr := cookie.Value

	token, err := jwt.ParseWithClaims(tokenStr, jwt.MapClaims{}, func(token *jwt.Token) (any, error) {
		return s.JwtPublic, nil
	})

	if err != nil {
		log.Println("No token")
		return "", errors.New("No token")
	}

	date, err := token.Claims.GetExpirationTime()

	if err != nil {
		log.Println("No expiration date")
		return "", errors.New("No expiration date")
	}

	if date.Time.Before(time.Now()) {
		log.Println("Expired token")
		return "", errors.New("Expired token")
	}

	iss, err := token.Claims.GetIssuer()

	if err != nil {
		return "", errors.New("Bad issuer")
	}

	return iss, nil
}

func (s *MainServiceHandler) Update(w http.ResponseWriter, req *http.Request) {

	if req.Method != http.MethodPost {
		log.Println("Wrong method in Update")
		http.Error(w, "Update is allowed only with POST method", http.StatusBadRequest)
		return
	}

	user, err := s.CheckToken(req)

	if err != nil {
		log.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	info, status, err := common.GetJsonStruct[common.UserInfo](req)

	if err != nil {
		log.Println(err.Error())
		http.Error(w, err.Error(), status)
		return
	}

	const query = "UPDATE Users SET first_name=$1, second_name=$2, date_of_birth=$3, email=$4, phone_number=$5 WHERE login=$6"

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	defer cancel()

	_, err = s.Db.ExecContext(ctx, query, info.FirstName, info.SecondName, info.DateOfBirth, info.Email, info.PhoneNumber, user)

	if err != nil {
		log.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (s *MainServiceHandler) CreatePost(w http.ResponseWriter, req *http.Request) {

	if req.Method != http.MethodPost {
		log.Println("Wrong method in CreatePost")
		http.Error(w, "Post method is one allowed", http.StatusBadRequest)
		return
	}

	user, err := s.CheckToken(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	info, status, err := common.GetJsonStruct[common.PostInfo](req)

	if err != nil {
		log.Println(err.Error())
		http.Error(w, err.Error(), status)
		return
	}

	id, err := s.client.NewPost(req.Context(), &pb.PostInfo{
		Author:           user,
		DateOfCreation:   info.DateOfCreation,
		Content:          info.Content,
		CommentSectionId: info.CommentSectionId,
	})

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = json.NewEncoder(w).Encode(id)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Println(err.Error())
		return
	}
	w.Header().Set("Content-Type", "application/json")
}

func (s *MainServiceHandler) UpdatePost(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		log.Println("Wrong method in UpdatePost")
		http.Error(w, "Post method is one allowed", http.StatusBadRequest)
		return
	}

	user, err := s.CheckToken(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	info, status, err := common.GetJsonStruct[common.PostInfo](req)

	if err != nil {
		log.Println(err.Error())
		http.Error(w, err.Error(), status)
		return
	}

	_, err = s.client.UpdatePost(req.Context(), &pb.PostInfo{
		Author:           user,
		DateOfCreation:   info.DateOfCreation,
		Content:          info.Content,
		CommentSectionId: info.CommentSectionId,
	})

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (s *MainServiceHandler) DeletePost(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		log.Println("Wrong method in DeletePost")
		http.Error(w, "Post method is one allowed", http.StatusBadRequest)
		return
	}

	user, err := s.CheckToken(req)
	if err != nil {
		http.Error(w, "No token?", http.StatusBadRequest)
		return
	}

	info, status, err := common.GetJsonStruct[common.PostId](req)

	if err != nil {
		log.Println(err.Error())
		http.Error(w, err.Error(), status)
		return
	}

	_, err = s.client.DeletePost(req.Context(), &pb.PostIdAuthor{Id: info.Id, Author: user})

	if err != nil {
		log.Println(err.Error())
		http.Error(w, err.Error(), status)
		return
	}
}

func (s *MainServiceHandler) GetPost(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		log.Println("Wrong method in GetPost")
		http.Error(w, "Post method is one allowed", http.StatusBadRequest)
		return
	}

	_, err := s.CheckToken(req)
	if err != nil {
		http.Error(w, "No token?", http.StatusBadRequest)
		return
	}

	info, status, err := common.GetJsonStruct[common.PostId](req)

	if err != nil {
		log.Println(err.Error())
		http.Error(w, err.Error(), status)
		return
	}

	res, err := s.client.GetPost(req.Context(), &pb.PostID{Id: info.Id})

	if err != nil {
		log.Println(err.Error())
		http.Error(w, err.Error(), status)
		return
	}

	err = json.NewEncoder(w).Encode(res)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Println(err.Error())
		return
	}
	w.Header().Set("Content-Type", "application/json")
}

func (s *MainServiceHandler) GetPostList(w http.ResponseWriter, req *http.Request) {

	if req.Method != http.MethodPost {
		log.Println("Wrong method in GetPostList")
		http.Error(w, "Post method is one allowed", http.StatusBadRequest)
		return
	}

	_, err := s.CheckToken(req)
	if err != nil {
		http.Error(w, "No token?", http.StatusBadRequest)
		return
	}

	info, status, err := common.GetJsonStruct[common.PaginationInfo](req)

	if err != nil {
		log.Println(err.Error())
		http.Error(w, err.Error(), status)
		return
	}

	res, err := s.client.GetPosts(req.Context(), &pb.PaginationInfo{
		PageNumber: info.PageNumber,
		BatchSize:  info.BatchSize,
	})

	if err != nil {
		log.Println(err.Error())
		http.Error(w, err.Error(), status)
		return
	}

	err = json.NewEncoder(w).Encode(res)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Println(err.Error())
		return
	}
	w.Header().Set("Content-Type", "application/json")
}

func (s *MainServiceHandler) SendView(w http.ResponseWriter, req *http.Request) {

	if req.Method != http.MethodPost {
		http.Error(w, "Post method is the one allowed", http.StatusBadRequest)
		log.Println("Wrong method in SendView")
		return
	}

	user, err := s.CheckToken(req)

	if err != nil {
		log.Println(err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	info, status, err := common.GetJsonStruct[common.ReactionInfo](req)

	if err != nil {
		log.Println(err.Error())
		http.Error(w, err.Error(), status)
		return
	}

	_, err = s.client.GetPost(req.Context(), &pb.PostID{Id: info.PostId})

	if err != nil {
		log.Println(err.Error())
		http.Error(w, err.Error(), status)
		return
	}

	// kafka

	data, err := json.Marshal(&common.Reaction{
		User:   user,
		Author: "unused",
		PostId: info.PostId,
	})

	if err != nil {
		log.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	topic_name := "topic-view"

	err = s.producer.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{
			Topic:     &topic_name,
			Partition: kafka.PartitionAny},
		Value: []byte(data)},
		nil, // delivery channel
	)

	if err != nil {
		log.Println(err.Error())
		log.Println("kafka view produce error")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (s *MainServiceHandler) SendLike(w http.ResponseWriter, req *http.Request) {

	if req.Method != http.MethodPost {
		http.Error(w, "Post method is the one allowed", http.StatusBadRequest)
		log.Println("Wrong method in SendLike")
		return
	}

	user, err := s.CheckToken(req)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	info, status, err := common.GetJsonStruct[common.ReactionInfo](req)

	if err != nil {
		log.Println(err.Error())
		http.Error(w, err.Error(), status)
		return
	}

	// get post Author from PostService
	post, err := s.client.GetPost(context.Background(), &pb.PostID{Id: info.PostId})

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// kafka

	data, err := json.Marshal(&common.Reaction{
		User:   user,
		Author: post.Author,
		PostId: info.PostId,
	})

	if err != nil {
		log.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	topic_name := "topic-like"

	err = s.producer.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{
			Topic:     &topic_name,
			Partition: kafka.PartitionAny},
		Value: []byte(data)},
		nil, // delivery channel
	)

	if err != nil {
		log.Println(err.Error())
		log.Println("kafka like produce error")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Println("all good")
}

func (s *MainServiceHandler) TotalActivity(w http.ResponseWriter, req *http.Request) {

	if req.Method != http.MethodPost {
		log.Println("Wrong method in Total activity")
		http.Error(w, "Post method is one allowed", http.StatusBadRequest)
		return
	}

	_, err := s.CheckToken(req)
	if err != nil {
		http.Error(w, "No token?", http.StatusBadRequest)
		return
	}

	info, status, err := common.GetJsonStruct[common.PostId](req)

	if err != nil {
		log.Println(err.Error())
		http.Error(w, err.Error(), status)
		return
	}

	log.Println("Total activity id", info.Id)

	res, err := s.stats.Total(req.Context(), &rpc_stats.PostID{Id: info.Id})

	if err != nil {
		log.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = json.NewEncoder(w).Encode(res)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Println(err.Error())
		return
	}
	w.Header().Set("Content-Type", "application/json")
}

func (s *MainServiceHandler) TopPosts(w http.ResponseWriter, req *http.Request) {

	if req.Method != http.MethodPost {
		log.Println("Wrong method in Top Posts")
		http.Error(w, "Post method is one allowed", http.StatusBadRequest)
		return
	}

	_, err := s.CheckToken(req)
	if err != nil {
		http.Error(w, "No token?", http.StatusBadRequest)
		return
	}

	info, status, err := common.GetJsonStruct[common.PostIsLike](req)

	if err != nil {
		log.Println(err.Error())
		http.Error(w, err.Error(), status)
		return
	}

	res, err := s.stats.Top(context.Background(), &rpc_stats.TopInfo{IsLike: info.IsLike})

	if err != nil {
		log.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// post authors

	for _, activity := range res.Ids {

		res, err := s.client.GetPost(req.Context(), &pb.PostID{Id: activity.Id})

		if err != nil {
			log.Println(err.Error())
			http.Error(w, err.Error(), status)
			return
		}

		activity.AuthorLogin = res.Author
	}

	err = json.NewEncoder(w).Encode(res)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Println(err.Error())
		return
	}
	w.Header().Set("Content-Type", "application/json")
}

func (s *MainServiceHandler) TopUsers(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		log.Println("Wrong method in Top Users")
		http.Error(w, "Post method is one allowed", http.StatusBadRequest)
		return
	}

	_, err := s.CheckToken(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	res, err := s.stats.Rating(context.Background(), &emptypb.Empty{})

	if err != nil {
		log.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = json.NewEncoder(w).Encode(res)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Println(err.Error())
		return
	}
	w.Header().Set("Content-Type", "application/json")
}
