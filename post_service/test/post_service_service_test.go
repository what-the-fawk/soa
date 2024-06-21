package post_service_test

import (
	"context"
	"os/exec"
	"soa/post_service/posts_service/pkg/pb"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func TestPostServiceRPC(t *testing.T) {

	// cmd := exec.Command("docker-compose", "up", "post_service", "post_db", "broker", "zookeeper")
	// go cmd.Run()

	// time.Sleep(time.Second * 90)

	conn, err := grpc.Dial("localhost:6666", grpc.WithTransportCredentials(insecure.NewCredentials()))

	if err != nil {
		t.Fatal(err)
	}

	defer conn.Close()

	client := pb.NewPostServiceClient(conn)

	req_create := &pb.PostInfo{
		Author:           "test_pixie",
		DateOfCreation:   "10.11.2012",
		Content:          "test content",
		CommentSectionId: 0,
	}

	post_id, err := client.NewPost(context.Background(), req_create)

	if err != nil {
		t.Fatal(err.Error())
	}

	post, err := client.GetPost(context.Background(), post_id)

	if err != nil {
		t.Fatal(err.Error())
	}

	if post.Author != req_create.Author {
		t.Fatal("Author incorrect")
	}

	if post.CommentSectionId != req_create.CommentSectionId {
		t.Fatal("Comment section ID incorrect")
	}

	if post.Content != req_create.Content {
		t.Fatal("Content incorrect")
	}

	if post.DateOfCreation != req_create.DateOfCreation {
		t.Fatal("Date of creation incorrect")
	}

	if post.Id != post_id.Id {
		t.Fatal("Incorrect PostID generation")
	}

	cmd2 := exec.Command("docker-compose", "down")
	cmd2.Run()
}
