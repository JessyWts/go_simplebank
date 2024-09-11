package api

import (
	"database/sql"
	"net/http"

	db "bitbucket.org/jessyw/go_simplebank/db/sqlc"
	"github.com/gin-gonic/gin"
)

type createEntryRequest struct {
	AccountID int64 `json:"account_id" binding:"required,min=1"`
	Amount    int64 `json:"amount" binding:"required"`
}

func (server *Server) CreateEntry(ctx *gin.Context) {
	var req createEntryRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	_, err := server.store.GetAccount(ctx, req.AccountID)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	arg := db.CreateEntryParams{
		AccountID: req.AccountID,
		Amount:    req.Amount,
	}

	entry, err := server.store.CreateEntry(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, entry)
}

type findEntryByAccountIDEntryRequest struct {
	ID int64 `uri:"id" binding:"required,min=1"`
}

// FindEntryByAccountID - find one entry from an account ID
func (server *Server) FindEntryByAccountID(ctx *gin.Context) {
	var req findEntryByAccountIDEntryRequest

	if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	entry, err := server.store.GetEntry(ctx, req.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, entry)
}

type getListEntriesByIdRequest struct {
	ID     int64 `form:"id" binding:"required,min=1"`
	Offset int32 `form:"offset" binding:"required,min=1"`
	Limit  int32 `form:"limit" binding:"required,min=5,max=10"`
}

// GetEntriesListById - get a list of entries from an account ID
func (server *Server) GetEntriesListById(ctx *gin.Context) {
	var req getListEntriesByIdRequest

	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	arg := db.ListEntriesParams{
		AccountID: req.ID,
		Limit:     req.Limit,
		Offset:    (req.Offset - 1) * req.Limit,
	}

	entries, err := server.store.ListEntries(ctx, arg)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse((err)))
		return
	}

	ctx.JSON(http.StatusOK, entries)
}
