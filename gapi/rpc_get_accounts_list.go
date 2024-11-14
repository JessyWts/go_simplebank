package gapi

import (
	"context"

	db "bitbucket.org/jessyw/go_simplebank/db/sqlc"
	"bitbucket.org/jessyw/go_simplebank/pb"
	"bitbucket.org/jessyw/go_simplebank/util"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (server *Server) GetAccountsList(ctx context.Context, req *pb.GetAccountsListRequest) (*pb.GetAccountsListResponse, error) {
	authPayload, err := server.authorizedUser(ctx, []string{util.BankerRole, util.AdminRole, util.DepositorRole})
	if err != nil {
		return nil, unauthenticatedError(err)
	}

	if req.GetLimit() <= 0 || req.GetLimit() > 100 || req.GetOffset() < 1 {
		return nil, status.Errorf(codes.Internal, "invalid parameters")
	}

	arg := db.ListAccountsParams{
		Owner:  authPayload.Username,
		Limit:  req.GetLimit(),
		Offset: (req.GetOffset() - 1) * req.GetLimit(),
	}
	accounts := []db.Account{}

	if authPayload.Username == "admin" {
		accounts, err = server.store.ListAllAccounts(ctx, db.ListAllAccountsParams{
			Limit:  req.GetLimit(),
			Offset: (req.GetOffset() - 1) * req.GetLimit(),
		})
	} else {
		accounts, err = server.store.ListAccounts(ctx, arg)
	}

	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get accounts: %s", err)
	}

	response := &pb.GetAccountsListResponse{
		Accounts: convertAccountsList(accounts),
	}

	return response, nil
}
