package social

import (
	"database/sql"
	"ddrp-relayer/store"

	"github.com/pkg/errors"
)

const (
	ModerationTypeLike = iota
	ModerationTypePin  = iota
)

func CreateLike(tx *sql.Tx, userID int, reference string) (*Moderation, error) {
	return createModeration(tx, ModerationTypeLike, userID, reference)
}

func CreatePin(tx *sql.Tx, userID int, reference string) (*Moderation, error) {
	return createModeration(tx, ModerationTypePin, userID, reference)
}

func GetModerationByID(querier store.Querier, id int) (*Moderation, error) {
	query := `
SELECT m.id, u.username, t.name, e.created_at, e.id, e.refhash, m.reference, m.moderation_type
FROM moderations m
JOIN envelopes e ON m.envelope_id = e.id
JOIN users u ON e.user_id = u.id
JOIN tlds t ON u.tld_id = t.id
WHERE m.id = $1
`
	moderation, err := scanModeration(querier.QueryRow(query, id))
	if err != nil {
		return nil, errors.Wrap(err, "error getting moderation by id")
	}
	return moderation, nil
}

func createModeration(tx *sql.Tx, modType int, userID int, reference string) (*Moderation, error) {
	if modType != ModerationTypeLike && modType != ModerationTypePin {
		panic("unknown modType")
	}
	wrapMsg := "error creating moderation"
	envelopeID, err := CreateEnvelope(tx, userID)
	if err != nil {
		return nil, errors.Wrap(err, wrapMsg)
	}
	query := `
INSERT INTO moderations (envelope_id, reference, moderation_type)
VALUES ($1, $2, $3)
RETURNING id
`
	var id int
	err = tx.QueryRow(query, envelopeID, reference, modType).Scan(&id)
	if err != nil {
		return nil, errors.Wrap(err, wrapMsg)
	}
	mod, err := GetModerationByID(tx, id)
	if err != nil {
		return nil, errors.Wrap(err, wrapMsg)
	}
	refhash, err := SetEnvelopeRefhash(tx, userID, envelopeID, mod.EnvelopeFormat())
	if err != nil {
		return nil, errors.Wrap(err, wrapMsg)
	}
	mod.Refhash = refhash
	return mod, nil
}

func scanModeration(row *sql.Row) (*Moderation, error) {
	moderation := new(Moderation)
	err := row.Scan(
		&moderation.ID,
		&moderation.Username,
		&moderation.TLD,
		&moderation.CreatedAt,
		&moderation.ID,
		&moderation.Refhash,
		&moderation.Reference,
		&moderation.Type,
	)
	if err != nil {
		return nil, errors.Wrap(err, "error scanning moderation")
	}
	return moderation, nil
}
