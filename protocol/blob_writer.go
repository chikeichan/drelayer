package protocol

import (
	"context"
	"github.com/pkg/errors"
	"io"
	"time"
)

const (
	BlobWriterMaxChunkSize = 1 * 1024 * 1024
	BlobSize               = 16 * 1024 * 1024
)

type BlobWriter struct {
	Client      DDRPClient
	Signer      Signer
	Name        string
	Truncate    bool
	writeClient DDRP_WriteClient
	txID        uint32
	offset      int
}

func NewBlobWriter(client DDRPClient, signer Signer, name string) *BlobWriter {
	return &BlobWriter{
		Client: client,
		Signer: signer,
		Name:   name,
	}
}

func (b *BlobWriter) Write(p []byte) (int, error) {
	if err := b.createWriteClient(); err != nil {
		return 0, err
	}

	toWrite, writeErr := b.WriteAt(p, int64(b.offset))
	b.offset += toWrite
	return toWrite, writeErr
}

func (b *BlobWriter) WriteAt(p []byte, off int64) (int, error) {
	if err := b.createWriteClient(); err != nil {
		return 0, err
	}

	toWrite := int64(len(p))
	if toWrite == 0 {
		return 0, nil
	}

	var writeErr error
	remaining := BlobSize - off
	if toWrite > remaining {
		writeErr = io.EOF
		toWrite = remaining
	}
	if toWrite > BlobWriterMaxChunkSize {
		writeErr = errors.New("chunk size too large")
		toWrite = BlobWriterMaxChunkSize
	}

	err := b.writeClient.Send(&WriteReq{
		TxID:   b.txID,
		Offset: uint32(off),
		Data:   p[:toWrite],
	})
	if err != nil {
		return 0, errors.Wrap(err, "failed to send write request")
	}
	return int(toWrite), writeErr
}

func (b *BlobWriter) CommitAndClose(broadcast bool) error {
	if b.writeClient == nil {
		return nil
	}

	if _, err := b.writeClient.CloseAndRecv(); err != nil {
		return errors.Wrap(err, "failed to close write stream")
	}
	b.writeClient = nil
	ctx := context.Background()
	precommitRes, err := b.Client.PreCommit(ctx, &PreCommitReq{
		TxID: b.txID,
	})
	if err != nil {
		return errors.Wrap(err, "failed to perform precommit")
	}
	ts := time.Now()
	var mr [32]byte
	copy(mr[:], precommitRes.MerkleRoot)
	sig, err := SignSeal(b.Signer, b.Name, ts, mr)
	if err != nil {
		return errors.Wrap(err, "failed to sign commitment")
	}
	_, err = b.Client.Commit(ctx, &CommitReq{
		TxID:      b.txID,
		Timestamp: uint64(ts.Unix()),
		Signature: sig[:],
		Broadcast: broadcast,
	})
	if err != nil {
		return errors.Wrap(err, "failed to commit blob")
	}
	return nil
}

func (b *BlobWriter) createWriteClient() error {
	if b.writeClient != nil {
		return nil
	}

	ctx := context.Background()
	checkoutRes, err := b.Client.Checkout(ctx, &CheckoutReq{
		Name: b.Name,
	})
	if err != nil {
		return errors.Wrap(err, "failed to check out blob")
	}

	if b.Truncate {
		_, err = b.Client.Truncate(ctx, &TruncateReq{
			TxID: checkoutRes.TxID,
		})
		if err != nil {
			return errors.Wrap(err, "failed to truncate blob")
		}
	}

	b.txID = checkoutRes.TxID
	wc, err := b.Client.Write(context.Background())
	if err != nil {
		return errors.Wrap(err, "failed to open write stream")
	}
	b.writeClient = wc
	return nil
}

func (b *BlobWriter) Close() error {
	if b.writeClient == nil {
		return nil
	}

	_, err := b.writeClient.CloseAndRecv()
	return err
}
