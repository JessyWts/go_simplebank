package api

import (
	db "bitbucket.org/jessyw/go_simplebank/db/sqlc"
	"bitbucket.org/jessyw/go_simplebank/util"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

type Server struct {
	config util.Config
	store  db.Store
	router *gin.Engine
}

// NewServer create a new HTTP server an setup routing.
func NewServer(config util.Config, store db.Store) (*Server, error) {
	server := &Server{
		config: config,
		store:  store,
	}

	server.setupRouter()
	return server, nil
}

func (server *Server) setupRouter() {
	router := gin.Default()

	binding.Validator.Engine().(*validator.Validate).RegisterValidation("currency", validCurrency)

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
