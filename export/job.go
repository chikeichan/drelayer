package export

import (
	"bytes"
	"database/sql"
	"ddrp-relayer/log"
	"ddrp-relayer/protocol"
	apiv1 "ddrp-relayer/protocol/v1"
	"ddrp-relayer/social"
	"ddrp-relayer/store"
	"ddrp-relayer/tlds"
	"fmt"
	"github.com/ddrp-org/dformats"
	"github.com/pkg/errors"
	"io"
	"io/ioutil"
	"os"
	"time"
)

const (
	MaxDataPerSubdomain = 65280
)

var logger = log.WithModule("export")

func ExportTLDs(db *sql.DB, client apiv1.DDRPv1Client, signer protocol.Signer) error {
	wrapMsg := "error exporting tld"
	var started bool
	err := store.WithTransaction(db, func(tx *sql.Tx) error {
		s, err := StartJob(tx)
		if err != nil {
			return err
		}
		started = s
		return nil
	})
	if err != nil {
		return errors.Wrap(err, wrapMsg)
	}
	if !started {
		return nil
	}

	stream, err := tlds.Stream(db)
	if err != nil {
		return errors.Wrap(err, wrapMsg)
	}
	defer stream.Close()
	for stream.Next() {
		start := time.Now()
		tld, err := stream.Value()
		if err != nil {
			return errors.Wrap(err, wrapMsg)
		}
		if err := WriteTLD(db, client, signer, tld.Name); err != nil {
			return errors.Wrap(err, wrapMsg)
		}
		logger.Info("exported TLD", "tld", tld.Name, "duration", time.Since(start))
	}
	err = store.WithTransaction(db, func(tx *sql.Tx) error {
		if err := EndJob(tx); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return errors.Wrap(err, wrapMsg)
	}
	return nil
}

func WriteTLD(querier store.Querier, client apiv1.DDRPv1Client, signer protocol.Signer, tld string) error {
	wrapMsg := "error writing TLD"
	f, err := FormatTLD(querier, tld)
	if err != nil {
		return errors.Wrap(err, wrapMsg)
	}
	defer f.Close()

	bw := protocol.NewBlobWriter(client, signer, tld)
	if _, err := io.CopyN(bw, f, protocol.BlobSize); err != nil {
		return errors.Wrap(err, wrapMsg)
	}
	if err := bw.CommitAndClose(true); err != nil {
		return errors.Wrap(err, wrapMsg)
	}
	return nil
}

func FormatTLD(querier store.Querier, tld string) (io.ReadCloser, error) {
	wrapMsg := "error formatting tld"
	subdomains, err := GetSubdomainsForTLD(querier, tld)
	if err != nil {
		return nil, errors.Wrap(err, wrapMsg)
	}
	tmp, err := ioutil.TempFile("", fmt.Sprintf("tldimport-%s-", tld))
	if err != nil {
		return nil, errors.Wrap(err, wrapMsg)
	}
	if err := tmp.Truncate(protocol.BlobSize); err != nil {
		return nil, errors.Wrap(err, wrapMsg)
	}
	if _, err := tmp.Write([]byte(dformats.SubdomainMagic)); err != nil {
		return nil, errors.Wrap(err, wrapMsg)
	}
	for _, sub := range subdomains {
		record := &dformats.SubdomainRecord{
			Name:  sub.Username,
			Index: sub.Index,
		}
		if err := record.Encode(tmp); err != nil {
			return nil, errors.Wrap(err, wrapMsg)
		}
		logger.Info("formatted subdomain record", "subdomain", sub.Username)
	}
	if _, err := tmp.Seek(dformats.EndReservedDataOffset, io.SeekStart); err != nil {
		return nil, errors.Wrap(err, wrapMsg)
	}
	for _, sub := range subdomains {
		if err := writeSubdomain(querier, tmp, sub); err != nil {
			return nil, errors.Wrap(err, wrapMsg)
		}
		logger.Info("formatted subdomain data", "subdomain", sub.Username)
	}
	// seek to beginning
	if _, err := tmp.Seek(0, io.SeekStart); err != nil {
		return nil, errors.Wrap(err, wrapMsg)
	}

	return tmp, nil
}

func writeSubdomain(querier store.Querier, tmp *os.File, sub *Subdomain) error {
	wrapMsg := "error formatting subdomain"
	stream, err := StreamMessageManifestsForUserID(querier, sub.UserID)
	if err != nil {
		return errors.Wrap(err, wrapMsg)
	}
	defer stream.Close()
	buf := new(bytes.Buffer)
	var written int
	for stream.Next() {
		manifest, err := stream.Value()
		if err != nil {
			return errors.Wrap(err, wrapMsg)
		}
		var formattable social.EnvelopeFormatter
		if manifest.PostID != 0 {
			formattable, err = social.GetPostByID(querier, manifest.PostID)
		} else if manifest.ConnectionID != 0 {
			formattable, err = social.GetConnectionByID(querier, manifest.ConnectionID)
		} else if manifest.ModerationID != 0 {
			formattable, err = social.GetModerationByID(querier, manifest.ModerationID)
		} else {
			return errors.Wrap(errors.New("manifest has no message candidates"), wrapMsg)
		}
		if err != nil {
			return errors.Wrap(err, wrapMsg)
		}
		if err := dformats.EncodeEnvelope(buf, formattable.EnvelopeFormat()); err != nil {
			return errors.Wrap(err, wrapMsg)
		}
		if written+buf.Len() > MaxDataPerSubdomain {
			break
		}
		n, err := tmp.Write(buf.Bytes())
		if err != nil {
			return errors.Wrap(err, wrapMsg)
		}
		written += n
		msgType := formattable.EnvelopeFormat().Message.Type()
		logger.Info("wrote message", "sub", sub.Username, "type", string(msgType[:]))
		buf.Reset()
	}
	return nil
}
