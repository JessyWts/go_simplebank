package db

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Store defines all functions to execute db queries and transactions
type Store struct {
	*Queries
	db DBTX
}

// NewStore creates a new Store
func NewStore(db DBTX) *Store {
	return &Store{
		db:      db,
		Queries: New(db),
	}
}

type StorePool struct {
	*Queries
	dbPool *pgxpool.Pool
}

// NewStore creates a new Store
func NewStorePool(db *pgxpool.Pool) *StorePool {
	return &StorePool{
		dbPool:  db,
		Queries: New(db),
	}
}

func (store *Store) execTx(ctx context.Context, fn func(*Queries) error) error {
	tx, err := store.db.Begin(ctx)
	if err != nil {
		return err
	}

	q := New(tx)
	err = fn(q)
	if err != nil {
		if rbErr := tx.Rollback(ctx); rbErr != nil {
			return fmt.Errorf("tx err: %v, rb err: %v", err, rbErr)
		}
		return err
	}

	return tx.Commit(ctx)
}

// func (store *StorePool) execTxPool(ctx context.Context, fn func(*Queries) error) error {
// 	// tx, err := store.db.Begin(ctx)
// 	tx, err := store.dbPool.Begin(ctx)
// 	if err != nil {
// 		return err
// 	}

// 	q := NewStorePool(tx)
// 	err = fn(q)
// 	if err != nil {
// 		if rbErr := tx.Rollback(ctx); rbErr != nil {
// 			return fmt.Errorf("tx err: %v, rb err: %v", err, rbErr)
// 		}
// 		return err
// 	}

// 	return tx.Commit(ctx)
// }

// TransferTxParams contains the input parameters of the transfer transaction
type TransferTxParams struct {
	FromAccountID int64 `json:"from_account_id"`
	ToAccountID   int64 `json:"to_account_id"`
	Amount        int64 `json:"amount"`
}

// TransferTxResult is the result of the transfer transaction
type TransferTxResult struct {
	Transfer    Transfer `json:"transfer"`
	FromAccount Account  `json:"from_account"`
	ToAccount   Account  `json:"to_account"`
	FromEntry   Entry    `json:"from_entry"`
	ToEntry     Entry    `json:"to_entry"`
}

var txKey = struct{}{}

// TransferTx performs a money transfer from one account to the other.
// it create a transfer record, add account entries, and update accounts balance within a single database transaction
func (store *Store) TransferTx(ctx context.Context, arg TransferTxParams) (TransferTxResult, error) {
	var result TransferTxResult

	err := store.execTx(ctx, func(q *Queries) error {
		var err error

		// create transfer record
		txName := ctx.Value(txKey)

		fmt.Println(txName, "create transfer")
		result.Transfer, err = q.CreateTransfer(ctx, CreateTransferParams{
			FromAccountID: arg.FromAccountID,
			ToAccountID:   arg.ToAccountID,
			Amount:        arg.Amount,
		})
		if err != nil {
			return err
		}

		// create from entry
		fmt.Println(txName, "create from entry 1")
		result.FromEntry, err = q.CreateEntry(ctx, CreateEntryParams{
			AccountID: arg.FromAccountID,
			Amount:    -arg.Amount,
		})
		if err != nil {

			return err
		}

		// create to entry
		fmt.Println(txName, "create to entry 2")
		result.ToEntry, err = q.CreateEntry(ctx, CreateEntryParams{
			AccountID: arg.ToAccountID,
			Amount:    -arg.Amount,
		})
		if err != nil {

			return err
		}

		if arg.FromAccountID < arg.ToAccountID {
			result.FromAccount, result.ToAccount, err = addMoney(ctx, q, arg.FromAccountID, -arg.Amount, arg.ToAccountID, arg.Amount)
			if err != nil {
				return err
			}
		} else {
			result.ToAccount, result.FromAccount, err = addMoney(ctx, q, arg.ToAccountID, arg.Amount, arg.FromAccountID, -arg.Amount)
			if err != nil {
				return err
			}
		}
		return nil
	})

	return result, err
}

func addMoney(ctx context.Context, q *Queries, accountID int64, amount int64, accountID2 int64, amount2 int64) (account Account, account2 Account, err error) {
	account, err = q.AddAccountBalance(ctx, AddAccountBalanceParams{
		ID:     accountID,
		Amount: amount,
	})
	if err != nil {
		return
	}

	account2, err = q.AddAccountBalance(ctx, AddAccountBalanceParams{
		ID:     accountID2,
		Amount: amount2,
	})
	if err != nil {
		return
	}
	return
}
