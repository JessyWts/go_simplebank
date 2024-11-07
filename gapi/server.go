package gapi

import (
	"fmt"

	db "bitbucket.org/jessyw/go_simplebank/db/sqlc"
	"bitbucket.org/jessyw/go_simplebank/pb"
	"bitbucket.org/jessyw/go_simplebank/token"
	"bitbucket.org/jessyw/go_simplebank/util"
	"bitbucket.org/jessyw/go_simplebank/worker"
)

// Server serve gRPC requests for our banking service.
type Server struct {
	pb.UnimplementedSimpleBankServer
	config          util.Config
	store           db.Store
	tokenMaker      token.Maker
	taskDistributor worker.TaskDistributor
}

// NewServer create a new gRPC server.
func NewServer(config util.Config, store db.Store, taskDistributor worker.TaskDistributor) (*Server, error) {
	tokenMaker, err := token.NewPasetoMaker(config.TokenSymmetricKey)
	if err != nil {
		return nil, fmt.Errorf("cannot create token maker: %w", err)
	}

	server := &Server{
		config:          config,
		store:           store,
		tokenMaker:      tokenMaker,
		taskDistributor: taskDistributor,
	}

	return server, nil
}
