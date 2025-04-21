package database

import (
	"database/sql"
	"github.com/pixie-sh/errors-go"
	"github.com/pixie-sh/logger-go/logger"
	"gorm.io/gorm"
	"runtime/debug"
)

type TxOptions = sql.TxOptions
type IsolationLevel = sql.IsolationLevel

const (
	IsolationLevelDefault IsolationLevel = iota
	IsolationLevelReadUncommitted
	IsolationLevelReadCommitted
	IsolationLevelWriteCommitted
	IsolationLevelRepeatableRead
	IsolationLevelSnapshot
	IsolationLevelSerializable
	IsolationLevelLinearizable
)

type Repository[T any] struct {
	*DB

	// newInstance is used when a copy of current repository is needed
	// example in WithTx method
	newInstance func(*DB) T

	// Panic is a function to be overwritten if the repository specialized wants to
	// handle errors from panics. eventually rethrow the panic
	// default: ignores the panic error
	Panic func(r any) error
}

func NewRepository[T any](db *DB, newInstance func(*DB) T) Repository[T] {
	return Repository[T]{
		db,
		newInstance,
		func(r any) error {
			var pErr error
			switch v := r.(type) {
			case error:
				pErr = v
			default:
				pErr = errors.New("unknown internal error; %v", v).WithErrorCode(errors.DBErrorCode)
			}

			logger.Logger.With("error", pErr).With("st", debug.Stack()).Error("transaction being recovered in f: %v", pErr)
			return pErr
		},
	}
}

// Tx open transaction
// Deprecated: replaced with Transaction override function
func (repo Repository[T]) Tx(f func(*DB) error, opts ...*TxOptions) error {
	return repo.Transaction(f, opts...)
}

func (repo Repository[T]) Transaction(f func(*DB) error, opts ...*TxOptions) (pErr error) {
	if repo.DB == nil {
		return errors.New("no connection available for transaction").WithErrorCode(errors.DBErrorCode)
	}

	tx := repo.DB.Begin(opts...)
	if tx.Error != nil {
		return errors.NewWithError(tx.Error, "unable to start transaction").WithErrorCode(errors.DBErrorCode)
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			pErr = repo.Panic(r)
		}
	}()

	if err := f(tx); err != nil {
		tx.Rollback()
		logger.Logger.With("error", err).Error("error during execution within transaction, rolled back")
		return err
	}

	if err := tx.Commit().Error; err != nil {
		return errors.NewWithError(err, "error during commit. not committed").WithErrorCode(errors.DBErrorCode)
	}

	return nil
}

// WithTx creates a copy of current repository with txDB *DB connection
// uses provided function to duplicate repository
func (repo Repository[T]) WithTx(txDB *DB) T {
	if txDB == nil {
		return repo.newInstance(repo.DB)
	}

	return repo.newInstance(txDB)
}

func (repo Repository[T]) UpdatesWithError(values interface{}) (*DB, error) {
	result := repo.DB.Updates(values)
	if result.Error != nil {
		return result, result.Error
	}

	if result.RowsAffected == 0 {
		return result, gorm.ErrRecordNotFound
	}

	return result, nil
}
