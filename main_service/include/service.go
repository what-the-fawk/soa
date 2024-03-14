package service

import (
	"context"
	"crypto/rsa"
	"crypto/sha256"
	"database/sql"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	_ "github.com/lib/pq"
	"log"
	"net/http"
	"soa/common"
	"time"
)

type MainServiceHandler struct {
	db         *sql.DB
	jwtPrivate *rsa.PrivateKey
	jwtPublic  *rsa.PublicKey
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
		"password VARCHAR (40) NOT NULL, " +
		"first_name VARCHAR (40) NOT NULL, " +
		"second_name VARCHAR (40) NOT NULL, " +
		"date_of_birth VARCHAR (50) NOT NULL, " +
		"email VARCHAR (40) NOT NULL," +
		"phone_number VARCHAR (40) NOT NULL" +
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

	return &MainServiceHandler{
		db:         db,
		jwtPublic:  pub,
		jwtPrivate: pri,
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

	info, status, err := common.GetJsonStruct[common.NewUserInfo](req)

	if err != nil {
		log.Println("Json unmarshall error")
		http.Error(w, err.Error(), status)
		return
	}

	const query = "" +
		"INSERT INTO Users " +
		"(login, password, first_name, second_name, date_of_birth, email, phone_number) " +
		"VALUES ($1, $2, $3, $4, $5, $6, $7)"

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	defer cancel()

	passwordHash := sha256.Sum256([]byte(info.Password))

	_, err = s.db.ExecContext(ctx, query, info.Login, passwordHash, info.FirstName, info.SecondName, info.DateOfBirth, info.Email, info.PhoneNumber)

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

	passwordHash := sha256.Sum256([]byte(info.Password))

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

func (s *MainServiceHandler) Update(w http.ResponseWriter, req *http.Request) {

	if req.Method != http.MethodPatch {
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
