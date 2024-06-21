package post_service_test

import (
	"context"
	"log"
	postservice "soa/post_service/include"
	"soa/post_service/posts_service/pkg/pb"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestCreate(t *testing.T) {

	testCases := []pb.PostInfo{
		{
			Author:           "test1",
			DateOfCreation:   "25.10.2010",
			Content:          "Imagine",
			CommentSectionId: 0,
		},
		{
			Author:           "test2",
			DateOfCreation:   "11.10.2022",
			Content:          "some",
			CommentSectionId: 0,
		},
	}

	for _, tc := range testCases {
		db, mock, err := sqlmock.New()

		if err != nil {
			log.Fatal(err.Error())
		}

		serv := &postservice.PostService{
			Db:      db,
			Counter: 0,
		}

		mock.ExpectExec("INSERT INTO").WithArgs(1, tc.Author, tc.DateOfCreation, tc.Content, tc.CommentSectionId).WillReturnResult(sqlmock.NewResult(1, 1))

		res, err := serv.NewPost(context.Background(), &tc)

		if err != nil {
			log.Fatal(err.Error())
		}

		assert.Equal(t, res.Id, uint64(1))

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %s", err)
		}
	}
}

func TestUpdateInfo(t *testing.T) {

	testCases := []pb.PostInfo{
		{
			Author:           "test1",
			DateOfCreation:   "25.10.2010",
			Content:          "Imagine",
			CommentSectionId: 0,
		},
		{
			Author:           "test2",
			DateOfCreation:   "11.10.2022",
			Content:          "some",
			CommentSectionId: 0,
		},
	}

	for _, tc := range testCases {
		db, mock, err := sqlmock.New()

		if err != nil {
			log.Fatal(err.Error())
		}

		serv := &postservice.PostService{
			Db:      db,
			Counter: 0,
		}

		mock.ExpectExec("UPDATE").WithArgs(tc.Content, tc.Author, tc.DateOfCreation, tc.CommentSectionId).WillReturnResult(sqlmock.NewResult(1, 1))

		_, err = serv.UpdatePost(context.Background(), &tc)

		if err != nil {
			log.Fatal(err.Error())
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %s", err)
		}
	}
}

func TestDelete(t *testing.T) {

	testCases := []pb.PostIdAuthor{
		{
			Author: "test1",
			Id:     1,
		},
		{
			Author: "test2",
			Id:     2,
		},
	}

	for _, tc := range testCases {
		db, mock, err := sqlmock.New()

		if err != nil {
			log.Fatal(err.Error())
		}

		serv := &postservice.PostService{
			Db:      db,
			Counter: 0,
		}

		mock.ExpectExec("DELETE").WithArgs(tc.Id, tc.Author).WillReturnResult(sqlmock.NewResult(1, 1))

		_, err = serv.DeletePost(context.Background(), &tc)

		if err != nil {
			log.Fatal(err.Error())
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %s", err)
		}
	}
}
