package gapi

import (
	"context"

	db "bitbucket.org/jessyw/go_simplebank/db/sqlc"
	"bitbucket.org/jessyw/go_simplebank/pb"
	"bitbucket.org/jessyw/go_simplebank/validator"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (server *Server) CreateAccount(ctx context.Context, req *pb.CreateAccountRequest) (*pb.CreateAccountResponse, error) {

	authPayload, err := server.authorizedUser(ctx)
	if err != nil {
		return nil, unauthenticatedError(err)
	}

	if err := validator.ValidateCurrency(req.GetCurrency()); err != nil {
		return nil, invalidArgumentError([]*errdetails.BadRequest_FieldViolation{fieldViolation("currency", err)})
	}

	arg := db.CreateAccountParams{
		Owner:    authPayload.Username,
		Currency: req.GetCurrency(),
		Balance:  0,
	}

	account, err := server.store.CreateAccount(ctx, arg)
	if err != nil {
		errCode := db.ErrorCode(err)
		if errCode == db.UniqueViolation || errCode == db.ForeignKeyViolation {
			return nil, status.Errorf(codes.AlreadyExists, "for this username an account with the same currency already exists: %s", err)

		}
		return nil, status.Errorf(codes.Internal, "failed to create account: %s", err)
	}

	response := &pb.CreateAccountResponse{
		Account: convertAccount(account),
	}
	return response, nil

}
