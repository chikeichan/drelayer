package social

import (
	"database/sql"
	"ddrp-relayer/store"

	"github.com/pkg/errors"
)

const (
	ConnectionTypeFollow = iota
	ConnectionTypeBlock
)

func CreateFollow(tx *sql.Tx, userID int, subdomain string, tld string) (*Connection, error) {
	return createConnection(tx, ConnectionTypeFollow, userID, subdomain, tld)
}

func CreateBlock(tx *sql.Tx, userID int, username string, tld string) (*Connection, error) {
	return createConnection(tx, ConnectionTypeBlock, userID, username, tld)
}

func GetConnectionByID(querier store.Querier, id int) (*Connection, error) {
	query := `
SELECT c.id, u.username, t.name, e.created_at, e.id, e.refhash, c.tld, c.subdomain, c.connection_type
FROM connections c
JOIN envelopes e on c.envelope_id = e.id
JOIN users u on e.user_id = u.id
JOIN tlds t on u.tld_id = t.id
WHERE c.id = $1
`
	connection, err := scanConnection(querier.QueryRow(query, id))
	if err != nil {
		return nil, errors.Wrap(err, "error getting connection by id")
	}
	return connection, nil
}

func createConnection(tx *sql.Tx, connType int, userID int, subdomain string, tld string) (*Connection, error) {
	if connType != ConnectionTypeFollow && connType != ConnectionTypeBlock {
		panic("unknown connType")
	}
	wrapMsg := "error creating connection"
	envelopeID, err := CreateEnvelope(tx, userID)
	if err != nil {
		return nil, errors.Wrap(err, wrapMsg)
	}
	query := `
INSERT INTO connections (envelope_id, tld, connection_type, subdomain)
VALUES ($1, $2, $3, $4)
RETURNING id
`
	var id int
	err = tx.QueryRow(query, envelopeID, tld, connType, subdomain).Scan(&id)
	if err != nil {
		return nil, errors.Wrap(err, wrapMsg)
	}
	conn, err := GetConnectionByID(tx, id)
	if err != nil {
		return nil, errors.Wrap(err, wrapMsg)
	}
	refhash, err := SetEnvelopeRefhash(tx, userID, envelopeID, conn.EnvelopeFormat())
	if err != nil {
		return nil, errors.Wrap(err, wrapMsg)
	}
	conn.Refhash = refhash
	return conn, nil
}

func scanConnection(row *sql.Row) (*Connection, error) {
	connection := new(Connection)
	err := row.Scan(
		&connection.ID,
		&connection.Username,
		&connection.TLD,
		&connection.CreatedAt,
		&connection.ID,
		&connection.Refhash,
		&connection.ConnecteeTLD,
		&connection.ConnecteeSubdomain,
		&connection.Type,
	)
	if err != nil {
		return nil, errors.Wrap(err, "error scanning connection")
	}
	return connection, nil
}
