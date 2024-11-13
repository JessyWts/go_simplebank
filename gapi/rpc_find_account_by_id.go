package gapi

import (
	"context"
	"database/sql"

	"bitbucket.org/jessyw/go_simplebank/pb"
	"bitbucket.org/jessyw/go_simplebank/util"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (server *Server) FindAccountById(ctx context.Context, req *pb.FindAccountByIdRequest) (*pb.FindAccountByIdResponse, error) {
	_, err := server.authorizedUser(ctx, []string{util.BankerRole, util.AdminRole, util.DepositorRole})
	if err != nil {
		return nil, unauthenticatedError(err)
	}

	if req.GetId() <= 0 {
		return nil, status.Errorf(codes.Internal, "id cant be 0")
	}

	account, err := server.store.GetAccount(ctx, req.GetId())
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, status.Errorf(codes.NotFound, "account not found: %s", err)
		}
		return nil, status.Errorf(codes.NotFound, "failed to find account: %s", err)
	}

	response := &pb.FindAccountByIdResponse{
		Account: convertAccount(account),
	}
	return response, nil
}
