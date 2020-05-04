package user

import (
	"crypto/rand"
	"database/sql"
	"ddrp-relayer/store"
	"encoding/hex"
	"io"
	"time"

	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"
)

const (
	HashCost       = 10
	MaxSubsPerBlob = 255
)

var (
	ErrInvalidPassword = errors.New("password is invalid")
)

func CreateUsernamePassword(tx *sql.Tx, username string, tld string, email string, password string) (*User, error) {
	wrapMsg := "error creating username/password user"
	if err := validateCreatingUser(tx, username, tld); err != nil {
		return nil, errors.Wrap(err, wrapMsg)
	}
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), HashCost)
	if err != nil {
		return nil, errors.Wrap(err, wrapMsg)
	}
	var id int
	query := `
INSERT INTO users (tld_id, username, email, hashed_password)
VALUES ((SELECT t.id FROM tlds t WHERE t.name = $1), $2, $3, $4)
RETURNING id
`
	err = tx.QueryRow(query, tld, username, store.StringOrNil(email), hex.EncodeToString(passwordHash)).Scan(&id)
	if err != nil {
		return nil, errors.Wrap(err, wrapMsg)
	}
	user, err := GetByID(tx, id)
	if err != nil {
		return nil, errors.Wrap(err, wrapMsg)
	}
	return user, nil
}

func GetByID(querier store.Querier, id int) (*User, error) {
	query := `
SELECT u.id, t.name, u.username, u.email, u.created_at, u.updated_at FROM users u 
JOIN tlds t ON u.tld_id = t.id
WHERE u.id = $1
`
	user, err := scanUser(querier.QueryRow(query, id))
	if err != nil {
		return nil, errors.Wrap(err, "error finding user by id")
	}
	return user, nil
}

func GetByAPIToken(querier store.Querier, token string) (*User, error) {
	query := `
SELECT u.id, t.name, u.username, u.email, u.created_at, u.updated_at FROM users u 
JOIN tlds t ON u.tld_id = t.id
WHERE u.api_token = $1 AND u.api_token_created_at > now() - INTERVAL '24 hours'`
	user, err := scanUser(querier.QueryRow(query, token))
	if err != nil {
		return nil, errors.Wrap(err, "error finding user by id")
	}
	return user, nil
}

func ExistsByUsernameTLD(querier store.Querier, username string, tld string) (bool, error) {
	query := `
SELECT EXISTS(SELECT 1 FROM users u JOIN tlds t ON u.tld_id = t.id WHERE u.username = $1 AND t.name = $2) 
`
	var exists bool
	if err := querier.QueryRow(query, username, tld).Scan(&exists); err != nil {
		return false, errors.Wrap(err, "error getting user existence")
	}
	return exists, nil
}

func TLDExists(querier store.Querier, tld string) (bool, error) {
	query := `
SELECT EXISTS(SELECT 1 FROM tlds WHERE name = $1)
`
	var exists bool
	if err := querier.QueryRow(query, tld).Scan(&exists); err != nil {
		return false, errors.Wrap(err, "error getting tld existence")
	}
	return exists, nil
}

func CountSubdomainsForTLD(querier store.Querier, tld string) (int, error) {
	query := `
SELECT COUNT(*) FROM users u JOIN tlds t ON u.tld_id = t.id WHERE t.name = $1
`
	var subCount int
	if err := querier.QueryRow(query, tld).Scan(&subCount); err != nil {
		return 0, errors.Wrap(err, "error counting subdomains for tld")
	}
	return subCount, nil
}

func Authenticate(tx *sql.Tx, username string, tld string, password string) (string, time.Time, error) {
	wrapMsg := "error authenticating user"
	query := `
SELECT u.id, u.hashed_password FROM users u 
JOIN tlds t ON u.tld_id = t.id
WHERE u.username = $1 AND t.name = $2
`
	var (
		id             int
		hashedPassword string
	)
	err := tx.QueryRow(query, username, tld).Scan(&id, &hashedPassword)
	if err != nil {
		return "", time.Time{}, errors.Wrap(err, wrapMsg)
	}
	hashedPWBytes, err := hex.DecodeString(hashedPassword)
	if err != nil {
		return "", time.Time{}, errors.Wrap(err, wrapMsg)
	}
	if err := bcrypt.CompareHashAndPassword(hashedPWBytes, []byte(password)); err != nil {
		return "", time.Time{}, errors.Wrap(ErrInvalidPassword, wrapMsg)
	}
	tokenB := make([]byte, 32, 32)
	if _, err := io.ReadFull(rand.Reader, tokenB); err != nil {
		panic("error reading randomness")
	}
	token := hex.EncodeToString(tokenB)
	now := time.Now()
	_, err = tx.Exec(
		"UPDATE users SET api_token = $1, api_token_created_at = $2, updated_at = $2 WHERE id = $3",
		token,
		now,
		id,
	)
	if err != nil {
		return "", time.Time{}, errors.Wrap(err, wrapMsg)
	}
	return token, now, nil
}

func validateCreatingUser(querier store.Querier, username string, tld string) error {
	wrapMsg := "creating user failed validation"
	tldExists, err := TLDExists(querier, tld)
	if err != nil {
		return errors.Wrap(err, wrapMsg)
	}
	if !tldExists {
		return errors.Wrap(errors.New("tld does not exist"), wrapMsg)
	}
	usernameExists, err := ExistsByUsernameTLD(querier, username, tld)
	if err != nil {
		return errors.Wrap(err, wrapMsg)
	}
	if usernameExists {
		return errors.Wrap(errors.New("username already taken"), wrapMsg)
	}
	subCount, err := CountSubdomainsForTLD(querier, tld)
	if err != nil {
		return errors.Wrap(err, wrapMsg)
	}
	if subCount >= MaxSubsPerBlob {
		return errors.Wrap(errors.New("tld at maximum capacity"), wrapMsg)
	}
	return nil
}

func scanUser(row *sql.Row) (*User, error) {
	var email sql.NullString
	user := new(User)
	err := row.Scan(
		&user.ID,
		&user.TLD,
		&user.Username,
		&email,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	user.Email = email.String
	return user, nil
}
