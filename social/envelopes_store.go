package social

import (
	"database/sql"
	"ddrp-relayer/user"
	"encoding/hex"
	"github.com/ddrp-org/dformats"
	"github.com/pkg/errors"
)

func CreateEnvelope(tx *sql.Tx, userID int) (int, error) {
	query := `
INSERT INTO envelopes(user_id, guid)
VALUES($1, $2)
RETURNING id
`
	var id int
	if err := tx.QueryRow(query, userID, hexGuid()).Scan(&id); err != nil {
		return 0, errors.Wrap(err, "error creating envelope")
	}
	return id, nil
}

func hexGuid() string {
	data := dformats.NewGUID()
	return hex.EncodeToString(data[:])
}

func SetEnvelopeRefhash(tx *sql.Tx, userID int, envelopeID int, envelope *dformats.Envelope) (string, error) {
	wrapMsg := "error setting envelope refhash"
	u, err := user.GetByID(tx, userID)
	if err != nil {
		return "", errors.Wrap(err, wrapMsg)
	}
	refhashB, err := dformats.HashEnvelope(envelope, u.Username, u.TLD)
	if err != nil {
		return "", errors.Wrap(err, wrapMsg)
	}
	refhash := hex.EncodeToString(refhashB[:])
	query := "UPDATE envelopes SET refhash = $1 WHERE id = $2"
	_, err = tx.Exec(query, refhash, envelopeID)
	if err != nil {
		return "", errors.Wrap(err, wrapMsg)
	}
	return refhash, nil
}
