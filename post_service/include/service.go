package post_service

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/golang/protobuf/ptypes/empty"
	_ "github.com/lib/pq"
	"log"
	"soa/common"
	pb "soa/post_service/posts_service/pkg/pb"
	"sync/atomic"
	"time"
)

type PostService struct {
	pb.UnimplementedPostServiceServer
	db      *sql.DB
	counter uint64 // initially zero
}

const dbname = "postgres"
const connectionStringPattern string = "host=%s port=%d user=%s password=%s dbname=%s sslmode=%s"

func NewPostService() *PostService {

	host, port, user, password, sslmode, err := common.GetPostgresParams()

	log.Println(host, port, user, password, sslmode)

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
		"CREATE TABLE IF NOT EXISTS Posts " +
		"(" +
		"post_id NUMERIC UNIQUE NOT NULL, " +
		"author_id NUMERIC NOT NULL, " +
		"date_of_creation VARCHAR (40) NOT NULL, " +
		"content TEXT NOT NULL, " +
		"comment_section_id NUMERIC NOT NULL" +
		")"

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	_, err = db.ExecContext(ctx, query)

	if err != nil {
		log.Fatal(err.Error())
	}

	return &PostService{
		db:      db,
		counter: 0,
	}
}

func (s *PostService) NewPost(ctx context.Context, post *pb.PostInfo) (*pb.PostID, error) {

	const query = "INSERT INTO posts (post_id, author_id, date_of_creation, content, comment_section_id)" +
		" VALUES ($1, $2, $3, $4, $5) "

	newId := atomic.AddUint64(&s.counter, 1)

	_, err := s.db.Exec(query, newId, post.AuthorId, post.DateOfCreation, post.Content, post.CommentSectionId)

	if err != nil {
		return nil, err
	}

	retId := &pb.PostID{Id: newId}
	return retId, nil
}

func (s *PostService) UpdatePost(ctx context.Context, info *pb.PostInfo) (*empty.Empty, error) {

	const query = "UPDATE Posts SET content=$1 WHERE author_id=$2 AND date_of_creation=$3 AND comment_section_id=$4 "

	_, err := s.db.Exec(query, info.Content, info.AuthorId, info.DateOfCreation, info.CommentSectionId)

	return &empty.Empty{}, err
}

func (s *PostService) DeletePost(ctx context.Context, id *pb.PostID) (*empty.Empty, error) {

	const query = "DELETE FROM Posts WHERE post_id=$1 "

	_, err := s.db.Exec(query, id.Id)

	return &empty.Empty{}, err
}

func (s *PostService) GetPost(ctx context.Context, id *pb.PostID) (*pb.Post, error) {

	const query = "SELECT post_id, author_id, date_of_creation, content, comment_section_id from Posts WHERE post_id=$1"

	row := s.db.QueryRow(query, id.Id)

	post := &pb.Post{}

	err := row.Scan(&post.Id, &post.AuthorId, &post.DateOfCreation, &post.Content, &post.CommentSectionId)

	if err != nil {
		log.Println("Row scan error", err.Error())
	}

	return post, err
}

func (s *PostService) GetPosts(info *pb.PaginationInfo, server pb.PostService_GetPostsServer) error {

	const query = "SELECT post_id, content, author_id FROM Posts WHERE LIMIT $2 OFFSET $3"

	rows, err := s.db.Query(query, info.BatchSize, info.PageNumber*uint64(info.BatchSize))

	defer rows.Close()

	var posts []*pb.Post
	for rows.Next() {
		var id, auth_id uint64
		var content string
		err := rows.Scan(&id, &content, &auth_id)
		if err != nil {
			return nil, err
		}
		posts = append(posts, &pb.Post{Id: id, Content: content, AuthorId: auth_id})
	}

	return err
}

func (s *PostService) mustEmbedUnimplementedPostServiceServer() {
	//TODO implement me
}
