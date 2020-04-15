package social

import (
	"database/sql"
	"ddrp-relayer/store"
	"github.com/lib/pq"
	"github.com/pkg/errors"
)

func CreatePost(tx *sql.Tx, userID int, body string, title string, reference string, topic string, tags []string) (*Post, error) {
	wrapMsg := "error creating post"
	envelopeID, err := CreateEnvelope(tx, userID)
	if err != nil {
		return nil, errors.Wrap(err, wrapMsg)
	}
	query := `
INSERT INTO posts (envelope_id, body, title, reference, topic, tags)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id
`
	var id int
	err = tx.QueryRow(
		query,
		envelopeID,
		body,
		store.StringOrNil(title),
		store.StringOrNil(reference),
		store.StringOrNil(topic),
		pq.Array(tags),
	).Scan(&id)
	if err != nil {
		return nil, errors.Wrap(err, wrapMsg)
	}
	post, err := GetPostByID(tx, id)
	if err != nil {
		return nil, errors.Wrap(err, wrapMsg)
	}
	refhash, err := SetEnvelopeRefhash(tx, userID, envelopeID, post.EnvelopeFormat())
	if err != nil {
		return nil, errors.Wrap(err, wrapMsg)
	}
	post.Refhash = refhash
	return post, nil
}

func GetPostByID(querier store.Querier, id int) (*Post, error) {
	query := `
SELECT p.id, u.username, t.name, e.created_at, e.guid, e.refhash, p.body, p.title, p.reference, p.topic, p.tags
FROM posts p
JOIN envelopes e on p.envelope_id = e.id
JOIN users u on e.user_id = u.id
JOIN tlds t on u.tld_id = t.id
WHERE p.id = $1
`
	post, err := scanPost(querier.QueryRow(query, id))
	if err != nil {
		return nil, errors.Wrap(err, "error getting post by id")
	}
	return post, nil
}

func scanPost(row *sql.Row) (*Post, error) {
	post := new(Post)
	var tags pq.StringArray
	err := row.Scan(
		&post.ID,
		&post.Username,
		&post.TLD,
		&post.CreatedAt,
		&post.GUID,
		&post.Refhash,
		&post.Body,
		&post.Title,
		&post.Reference,
		&post.Topic,
		&tags,
	)
	if err != nil {
		return nil, errors.Wrap(err, "error scanning post")
	}
	post.Tags = tags
	return post, nil
}
