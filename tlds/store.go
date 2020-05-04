package tlds

import (
	"database/sql"
	"ddrp-relayer/protocol"
	"ddrp-relayer/store"

	"github.com/pkg/errors"
)

func Upsert(db *sql.DB, tlds []protocol.KeysNames) error {
	return store.WithTransaction(db, func(tx *sql.Tx) error {
		for _, tld := range tlds {
			_, err := tx.Exec("INSERT INTO tlds (name) VALUES ($1) ON CONFLICT DO NOTHING", tld.Name)
			if err != nil {
				return errors.Wrap(err, "error upserting TLD")
			}
		}
		return nil
	})
}

type TLDStream struct {
	rows *sql.Rows
}

func (t *TLDStream) Next() bool {
	return t.rows.Next()
}

func (t *TLDStream) Value() (*TLD, error) {
	tld := new(TLD)
	if err := t.rows.Scan(&tld.ID, &tld.Name); err != nil {
		return nil, errors.Wrap(err, "error streaming tld")
	}
	return tld, nil
}

func (t *TLDStream) Close() error {
	defer t.rows.Close()
	if err := t.rows.Err(); err != nil {
		return errors.Wrap(err, "error closing tld stream")
	}
	return nil
}

func Stream(querier store.Querier) (*TLDStream, error) {
	query := `
SELECT id, name FROM tlds ORDER BY name
`
	rows, err := querier.Query(query)
	if err != nil {
		return nil, errors.Wrap(err, "error opening tld stream")
	}
	return &TLDStream{
		rows: rows,
	}, nil
}
