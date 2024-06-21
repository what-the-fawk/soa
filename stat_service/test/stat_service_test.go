package main

import (
	"context"
	"log"
	statservice "soa/stat_service/include"
	"soa/stat_service/stats_service/pkg/pb"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"google.golang.org/protobuf/types/known/emptypb"
)

func TestTop(t *testing.T) {

	testCases := []struct {
		info    pb.TopInfo
		post_id uint64
		count   uint64
	}{
		{
			info:    pb.TopInfo{IsLike: 1},
			post_id: 10,
			count:   2,
		},
		{
			info:    pb.TopInfo{IsLike: 0},
			post_id: 32,
			count:   5,
		},
	}

	for _, tc := range testCases {
		db, mock, err := sqlmock.New()

		if err != nil {
			log.Fatal(err.Error())
		}

		serv := &statservice.StatServiceHandler{
			Click: db,
		}

		mock.ExpectBegin()
		mock.ExpectQuery("SELECT").WillReturnRows(sqlmock.NewRows([]string{"post_id", "`value_occurence`"}).AddRow(tc.post_id, tc.count))
		mock.ExpectCommit()

		_, err = serv.Top(context.Background(), &tc.info)

		if err != nil {
			log.Fatal(err.Error())
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %s", err)
		}
	}
}

func TestRating(t *testing.T) {
	testCases := []struct {
		info   pb.TopInfo
		author string
		count  uint64
	}{
		{
			info:   pb.TopInfo{IsLike: 1},
			author: "test1",
			count:  2,
		},
		{
			info:   pb.TopInfo{IsLike: 0},
			author: "test2",
			count:  5,
		},
	}

	for _, tc := range testCases {
		db, mock, err := sqlmock.New()

		if err != nil {
			log.Fatal(err.Error())
		}

		serv := &statservice.StatServiceHandler{
			Click: db,
		}

		mock.ExpectBegin()
		mock.ExpectQuery("SELECT").WillReturnRows(sqlmock.NewRows([]string{"post_id", "`value_occurence`"}).AddRow(tc.author, tc.count))
		mock.ExpectCommit()

		_, err = serv.Rating(context.Background(), &emptypb.Empty{})

		if err != nil {
			log.Fatal(err.Error())
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %s", err)
		}
	}
}

func TestActivity(t *testing.T) {
	testCases := []struct {
		info    pb.PostID
		post_id uint64
		likes   uint64
		views   uint64
	}{
		{
			info:  pb.PostID{Id: 1},
			likes: 2,
			views: 4,
		},
		{
			info:  pb.PostID{Id: 2},
			likes: 10,
			views: 3,
		},
	}

	for _, tc := range testCases {
		db, mock, err := sqlmock.New()

		if err != nil {
			log.Fatal(err.Error())
		}

		serv := &statservice.StatServiceHandler{
			Click: db,
		}

		mock.ExpectBegin()
		// likes
		mock.ExpectQuery("SELECT").WithArgs(tc.info.Id).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(tc.likes))
		//views
		mock.ExpectQuery("SELECT").WithArgs(tc.info.Id).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(tc.views))
		mock.ExpectCommit()

		_, err = serv.Total(context.Background(), &tc.info)

		if err != nil {
			log.Fatal(err.Error())
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %s", err)
		}
	}
}
