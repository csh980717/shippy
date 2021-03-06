package main

import (
	"errors"
	pb "github.com/csh980717/shippy/user-service/proto/auth"
	"github.com/micro/go-micro"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/net/context"
	"log"
)

type service struct {
	repo         Repository
	tokenService Authable
	pubSub       micro.Publisher
}

const topic = "auth.created"

func (s *service) Create(ctx context.Context, req *pb.User, res *pb.Response) error {
	log.Println("Creating user: ", req)
	hashedPass, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	req.Password = string(hashedPass)
	if err := s.repo.Create(req); err != nil {
		return err
	}
	res.User = req
	if err := s.pubSub.Publish(ctx, req); err != nil {
		return err
	}
	return nil
}

func (s *service) Get(ctx context.Context, req *pb.User, res *pb.Response) error {
	user, err := s.repo.Get(req.Id)
	if err != nil {
		return err
	}
	res.User = user
	return nil
}

func (s *service) GetAll(ctx context.Context, req *pb.Request, res *pb.Response) error {
	users, err := s.repo.GetAll()
	if err != nil {
		return err
	}
	res.Users = users
	return nil
}

func (s *service) Auth(ctx context.Context, req *pb.User, res *pb.Token) error {
	log.Println("Logging in with:", req.Email, req.Password)
	user, err := s.repo.GetByEmail(req.Email)
	log.Println(user, err)
	if err != nil {
		return err
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return err
	}
	token, err := s.tokenService.Encode(user)
	if err != nil {
		return err
	}
	res.Token = token
	return nil
}

func (s *service) ValidateToken(context context.Context, req *pb.Token, res *pb.Token) error {
	claims, err := s.tokenService.Decode(req.Token)
	if err != nil {
		return err
	}
	if claims.User.Id == "" {
		return errors.New("invalid user")
	}
	res.Valid = true
	return nil
}
