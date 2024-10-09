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

	router.POST("/login", server.LoginUser)
	router.POST("/users", server.CreateUser)

	authRoutes := router.Group("/").Use(AuthMiddleware(server.tokenMaker))

	authRoutes.GET("/users/:name", server.FindUserByName)

	authRoutes.POST("/accounts", server.CreateAccount)
	authRoutes.GET("/accounts/:id", server.FindAccountById)
	authRoutes.GET("/accounts", server.GetAccounts)
	authRoutes.DELETE("/accounts/:id", server.DeleteAccount)

	authRoutes.POST("/entries", server.CreateEntry)
	authRoutes.GET("/entries/:id", server.FindEntryByAccountID)
	authRoutes.GET("/entries", server.GetEntriesListById)

	authRoutes.POST("/transfers", server.CreateTransfert)

	server.router = router
}

// Start runs the HTTP server on a specific address.
func (server *Server) Start(address string) error {
	return server.router.Run(address)
}

func errorResponse(err error) gin.H {
	return gin.H{"error": err.Error()}
}
