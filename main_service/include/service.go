package service

import (
	"context"
	"crypto/rsa"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	_ "github.com/lib/pq"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"net/http"
	"os"
	"soa/common"
	"soa/post_service/posts_service/pkg/pb"
	"time"
)

type MainServiceHandler struct {
	db         *sql.DB
	jwtPrivate *rsa.PrivateKey
	jwtPublic  *rsa.PublicKey
	client     pb.PostServiceClient
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

	grpcServerAddr, ok := os.LookupEnv("GRPC_SERVER")
	if !ok {
		log.Fatalf("GRPC_SERVER not set")
	}
	conn, err := grpc.Dial(grpcServerAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect to gRPC server: %v", err)
	}
	grpcClient := pb.NewPostServiceClient(conn)

	return &MainServiceHandler{
		db:         db,
		jwtPublic:  pub,
		jwtPrivate: pri,
		client:     grpcClient,
	}
}

func (s *MainServiceHandler) Close() {
	s.db.Close()
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

	_, err = s.db.ExecContext(ctx, query, info.Login, passwordHash)

	if err != nil {
		log.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (s *MainServiceHandler) Auth(w http.ResponseWriter, req *http.Request) {

	if req.Method != http.MethodGet {
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

	row := s.db.QueryRowContext(ctx, query, info.Login)

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
		"iss": "MainService",
		"exp": 60 * time.Minute,
		"aud": userQueryInfo.Login,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)

	tokenStr, err := token.SignedString(s.jwtPrivate)

	http.SetCookie(w, &http.Cookie{
		Name:  "jwt",
		Value: tokenStr,
	})
}

func (s *MainServiceHandler) CheckToken(req *http.Request) error {
	return nil ////////////////////////////////////////////////////////////////
}

func (s *MainServiceHandler) Update(w http.ResponseWriter, req *http.Request) {

	if req.Method != http.MethodPut {
		log.Println("Wrong method in Update")
		http.Error(w, "Update is allowed only with POST method", http.StatusBadRequest)
		return
	}

	info, status, err := common.GetJsonStruct[common.UserInfo](req)

	if err != nil {
		log.Println(err.Error())
		http.Error(w, err.Error(), status)
		return
	}

	// check token

	cookie, err := req.Cookie("jwt")

	if err != nil {
		log.Println("No jwt?")
		http.Error(w, "Not authorized", http.StatusUnauthorized)
		return
	}

	tokenStr := cookie.Value

	token, err := jwt.ParseWithClaims(tokenStr, jwt.MapClaims{}, func(token *jwt.Token) (any, error) {
		return s.jwtPublic, nil
	})

	if err != nil {
		log.Println("No token")
		http.Error(w, "Invalid auth", http.StatusBadRequest)
		return
	}

	date, err := token.Claims.GetExpirationTime()

	if err != nil {
		log.Println("No expiration date")
		http.Error(w, "Invalid token", http.StatusBadRequest)
		return
	}

	if time.Now().Second() > date.Time.Second() {
		log.Println("Expired token")
		http.Error(w, "Expired token", http.StatusUnauthorized)
	}

	if !token.Valid {
		log.Println("Invalid token")
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	iss, err := token.Claims.GetIssuer()

	if err != nil || iss != "MainService" {
		http.Error(w, "Invalid auth", http.StatusBadRequest)
		return
	}

	login, err := token.Claims.GetAudience()

	if err != nil {
		log.Println("Invalid aud")
		http.Error(w, "Invalid auth", http.StatusBadRequest)
		return
	}

	const query = "UPDATE Users SET first_name=$1, second_name=$2, date_of_birth=$3, email=$4, phone_number=$5 WHERE login=$6"

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	defer cancel()

	_, err = s.db.ExecContext(ctx, query, info.FirstName, info.SecondName, info.DateOfBirth, info.Email, info.PhoneNumber, login)

	if err != nil {
		log.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

}

func (s *MainServiceHandler) CreatePost(w http.ResponseWriter, req *http.Request) {

	if req.Method != http.MethodPost {
		if req.Method != http.MethodGet {
			log.Println("Wrong method in CreatePost")
			http.Error(w, "Post method is one allowed", http.StatusBadRequest)
			return
		}
	}

	err := s.CheckToken(req)
	if err != nil {
		http.Error(w, "No token?", http.StatusBadRequest)
		return
	}

	info, status, err := common.GetJsonStruct[common.PostInfo](req)

	if err != nil {
		log.Println(err.Error())
		http.Error(w, err.Error(), status)
		return
	}

	id, err := s.client.NewPost(req.Context(), &pb.PostInfo{
		AuthorId:         info.AuthorId,
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
		if req.Method != http.MethodGet {
			log.Println("Wrong method in UpdatePost")
			http.Error(w, "Post method is one allowed", http.StatusBadRequest)
			return
		}
	}

	err := s.CheckToken(req)
	if err != nil {
		http.Error(w, "No token?", http.StatusBadRequest)
		return
	}

	info, status, err := common.GetJsonStruct[common.PostInfo](req)

	if err != nil {
		log.Println(err.Error())
		http.Error(w, err.Error(), status)
		return
	}

	_, err = s.client.UpdatePost(req.Context(), &pb.PostInfo{
		AuthorId:         info.AuthorId,
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
		if req.Method != http.MethodGet {
			log.Println("Wrong method in DeletePost")
			http.Error(w, "Post method is one allowed", http.StatusBadRequest)
			return
		}
	}

	err := s.CheckToken(req)
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

	_, err = s.client.DeletePost(req.Context(), &pb.PostID{Id: info.Id})

	if err != nil {
		log.Println(err.Error())
		http.Error(w, err.Error(), status)
		return
	}
}

func (s *MainServiceHandler) GetPost(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		if req.Method != http.MethodGet {
			log.Println("Wrong method in GetPost")
			http.Error(w, "Post method is one allowed", http.StatusBadRequest)
			return
		}
	}

	err := s.CheckToken(req)
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
	if req.Method != http.MethodGet {
		if req.Method != http.MethodGet {
			log.Println("Wrong method in GetPostList")
			http.Error(w, "Post method is one allowed", http.StatusBadRequest)
			return
		}
	}

	err := s.CheckToken(req)
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
