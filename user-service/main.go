package main

import (
	pb "github.com/csh980717/shippy/user-service/proto/auth"
	"github.com/micro/go-micro"
	_ "github.com/micro/go-plugins/registry/mdns"
	"log"
)

func main() {
	db, err := CreateConnection()
	defer db.Close()
	if err != nil {
		log.Fatalf("Could not connect to DB: %v", err)
	}
	db.AutoMigrate(&pb.User{})

	repo := &UserRepository{db: db}
	tokenService := &TokenService{repo: repo}
	s := micro.NewService(
		micro.Name("shippy.auth"))
	s.Init()
	publisher := micro.NewPublisher(topic, s.Client())
	pb.RegisterAuthHandler(s.Server(), &service{
		repo:         repo,
		tokenService: tokenService,
		pubSub:       publisher,
	})
	if err := s.Run(); err != nil {
		log.Fatal(err)
	}
}
