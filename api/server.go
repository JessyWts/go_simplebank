package api

import (
	"fmt"

	db "bitbucket.org/jessyw/go_simplebank/db/sqlc"
	"bitbucket.org/jessyw/go_simplebank/token"
	"bitbucket.org/jessyw/go_simplebank/util"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

type Server struct {
	config     util.Config
	store      db.Store
	tokenMaker token.Maker
	router     *gin.Engine
}

// NewServer create a new HTTP server an setup routing.
func NewServer(config util.Config, store db.Store) (*Server, error) {
	tokenMaker, err := token.NewPasetoMaker(config.TokenSymmetricKey)
	if err != nil {
		return nil, fmt.Errorf("cannot create token maker: %w", err)
	}

	server := &Server{
		config:     config,
		store:      store,
		tokenMaker: tokenMaker,
	}

	server.setupRouter()
	return server, nil
}

func (server *Server) setupRouter() {
	router := gin.Default()

	binding.Validator.Engine().(*validator.Validate).RegisterValidation("currency", validCurrency)

	router.POST("/users/login", server.LoginUser)
	router.POST("/users", server.CreateUser)
	router.GET("/users/:name", server.FindUserByName)

	router.POST("/accounts", server.CreateAccount)
	router.GET("/accounts/:id", server.FindAccountById)
	router.GET("/accounts", server.GetAccounts)
	router.DELETE("/accounts/:id", server.DeleteAccount)

	router.POST("/entries", server.CreateEntry)
	router.GET("/entries/:id", server.FindEntryByAccountID)
	router.GET("/entries", server.GetEntriesListById)

	router.POST("/transfers", server.CreateTransfert)

	server.router = router
}

// Start runs the HTTP server on a specific address.
func (server *Server) Start(address string) error {
	return server.router.Run(address)
}

func errorResponse(err error) gin.H {
	return gin.H{"error": err.Error()}
}
