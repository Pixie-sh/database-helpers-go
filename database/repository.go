package database

import (
	"github.com/pixie-sh/errors-go"
	"github.com/pixie-sh/logger-go/logger"
)

type Repository struct {
	*DB
}

func (repo Repository) Tx(f func(*DB) error) error {
	if repo.DB == nil {
		return errors.New("no connection available for transaction").WithErrorCode(errors.DBErrorCode)
	}

	tx := repo.DB.Begin()
	if tx.Error != nil {
		return errors.NewWithError(tx.Error, "unable to start transaction").WithErrorCode(errors.DBErrorCode)
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			logger.Logger.Error("transaction being recovered in f: %v", r)
		}
	}()

	if err := f(tx); err != nil {
		tx.Rollback()
		return errors.NewWithError(err, "error during execution within transaction, rolled back").WithErrorCode(errors.DBErrorCode)
	}

	if err := tx.Commit().Error; err != nil {
		return errors.NewWithError(err, "error during commit. not committed").WithErrorCode(errors.DBErrorCode)
	}

	return nil
}
