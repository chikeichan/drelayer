package store

import (
	"database/sql"
	"github.com/pkg/errors"
)

type TxCb func(tx *sql.Tx) error

func WithTransaction(db *sql.DB, cb TxCb) error {
	tx, err := db.Begin()
	if err != nil {
		return errors.Wrap(err, "failed to open transaction")
	}

	defer func() {
		if p := recover(); p != nil {
			// a panic occurred, rollback and repanic
			tx.Rollback()
			panic(p)
		} else if err != nil {
			// something went wrong, rollback
			tx.Rollback()
		} else {
			// all good, commit
			err = tx.Commit()
		}
	}()

	return cb(tx)
}

func StringOrNil(in string) interface{} {
	if in == "" {
		return nil
	}
	return in
}

func NilOrString(in *string) string {
	if in == nil {
		return ""
	}
	return *in
}
