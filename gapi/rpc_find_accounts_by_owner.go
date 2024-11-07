package gapi

import (
	"context"
	"database/sql"

	"bitbucket.org/jessyw/go_simplebank/pb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (server *Server) FindAccountsByOwner(ctx context.Context, req *pb.FindAccountsByOwnerRequest) (*pb.FindAccountsByOwnerResponse, error) {

	authPayload, err := server.authorizedUser(ctx)
	if err != nil {
		return nil, unauthenticatedError(err)
	}

	if req.GetOwner() == "" {
		return nil, status.Errorf(codes.Internal, "owner cant be empty")
	}

	if authPayload.Username != req.GetOwner() && authPayload.Username != "admin" {
		return nil, status.Errorf(codes.PermissionDenied, "you dont have the right to see this account")
	}

	accounts, err := server.store.GetListAccountsByOwner(ctx, req.GetOwner())
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, status.Errorf(codes.NotFound, "account not found: %s", err)
		}
		return nil, status.Errorf(codes.Internal, "failed to find account: %s", err)
	}

	response := &pb.FindAccountsByOwnerResponse{
		Accounts: convertAccountsList(accounts),
	}

	return response, nil
}
