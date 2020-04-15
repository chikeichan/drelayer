package export

import (
	"database/sql"
	"ddrp-relayer/store"
	"github.com/pkg/errors"
)

type Subdomain struct {
	UserID   int
	Username string
	Index    uint8
}

func GetSubdomainsForTLD(querier store.Querier, tld string) ([]*Subdomain, error) {
	wrapMsg := "error getting subdomains for tld"
	query := `
SELECT u.id, u.username FROM users u JOIN tlds t ON t.id = u.tld_id WHERE t.name = $1 ORDER BY u.username
`
	rows, err := querier.Query(query, tld)
	if err != nil {
		return nil, errors.Wrap(err, wrapMsg)
	}
	defer rows.Close()
	var subs []*Subdomain
	var i uint8
	for rows.Next() {
		sub := &Subdomain{
			Index: i,
		}
		if err := rows.Scan(&sub.UserID, &sub.Username); err != nil {
			return nil, errors.Wrap(err, wrapMsg)
		}
		subs = append(subs, sub)
		i++
	}
	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, wrapMsg)
	}
	return subs, nil
}

type MessageManifest struct {
	EnvelopeID   int
	PostID       int
	ConnectionID int
	ModerationID int
}

type MessageManifestStream struct {
	rows *sql.Rows
}

func (m *MessageManifestStream) Next() bool {
	return m.rows.Next()
}

func (m *MessageManifestStream) Value() (*MessageManifest, error) {
	var envelopeID int
	var postID sql.NullInt64
	var connectionID sql.NullInt64
	var moderationID sql.NullInt64
	if err := m.rows.Scan(&envelopeID, &postID, &connectionID, &moderationID); err != nil {
		return nil, errors.Wrap(err, "error streaming message manifest")
	}
	return &MessageManifest{
		EnvelopeID:   envelopeID,
		PostID:       int(postID.Int64),
		ConnectionID: int(connectionID.Int64),
		ModerationID: int(moderationID.Int64),
	}, nil
}

func (m *MessageManifestStream) Close() error {
	defer m.rows.Close()
	if err := m.rows.Err(); err != nil {
		return errors.Wrap(err, "error closing message manifest stream")
	}
	return nil
}

func StreamMessageManifestsForUserID(querier store.Querier, userID int) (*MessageManifestStream, error) {
	query := `
SELECT e.id, p.id AS post_id, c.id AS connection_id, m.id AS moderation_id FROM envelopes e
LEFT JOIN posts p ON p.envelope_id = e.id
LEFT JOIN connections c ON c.envelope_id = e.id
LEFT JOIN moderations m ON m.envelope_id = e.id
WHERE e.user_id = $1
ORDER BY created_at DESC
`
	rows, err := querier.Query(query, userID)
	if err != nil {
		return nil, errors.Wrap(err, "error opening message manifest stream")
	}
	return &MessageManifestStream{
		rows: rows,
	}, nil
}

func StartJob(tx *sql.Tx) (bool, error) {
	query := `
INSERT INTO ingestion_jobs (started_at) SELECT now() WHERE NOT EXISTS (SELECT 1 FROM ingestion_jobs WHERE ended_at IS NULL) RETURNING id
`
	var id sql.NullInt64
	err := tx.QueryRow(query).Scan(&id)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, errors.Wrap(err, "error starting job")
	}
	return id.Valid, nil
}

func EndJob(tx *sql.Tx) error {
	query := `
UPDATE ingestion_jobs SET ended_at = now() WHERE ended_at IS NULL
`
	_, err := tx.Exec(query)
	if err != nil {
		return errors.Wrap(err, "error ending job")
	}
	return nil
}
