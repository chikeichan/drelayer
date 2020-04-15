package protocol

import (
	"encoding/hex"
	"github.com/btcsuite/btcd/btcec"
	"github.com/mslipper/mstream"
	"github.com/pkg/errors"
	"golang.org/x/crypto/blake2b"
	"time"
)

type Signer interface {
	Sign(tld string, hash [32]byte) ([65]byte, error)
}

type KeysNames struct {
	Name string
	Key  string
}

type NameSigner struct {
	keys map[string]*btcec.PrivateKey
}

func NewNameSigner(kns []KeysNames) (Signer, error) {
	keys := make(map[string]*btcec.PrivateKey)

	for _, kn := range kns {
		keyB, err := hex.DecodeString(kn.Key)
		if err != nil {
			return nil, errors.Wrap(err, "invalid key hex")
		}
		priv, _ := btcec.PrivKeyFromBytes(btcec.S256(), keyB)
		keys[kn.Name] = priv
	}

	return &NameSigner{
		keys: keys,
	}, nil
}

func (n *NameSigner) Sign(tld string, hash [32]byte) ([65]byte, error) {
	var sig [65]byte
	pk := n.keys[tld]
	if pk == nil {
		return sig, errors.New("no private key for name")
	}

	sigBuf, err := btcec.SignCompact(btcec.S256(), pk, hash[:], false)
	if err != nil {
		return sig, err
	}
	copy(sig[:], sigBuf)
	return sig, nil
}

func SealHash(name string, ts time.Time, mr [32]byte) [32]byte {
	h, _ := blake2b.New256(nil)
	h.Write([]byte("DDRPBLOB"))
	if err := mstream.EncodeField(h, name); err != nil {
		panic(err)
	}
	if err := mstream.EncodeField(h, ts); err != nil {
		panic(err)
	}
	if _, err := h.Write(mr[:]); err != nil {
		panic(err)
	}
	var reservedRoot [32]byte
	if _, err := h.Write(reservedRoot[:]); err != nil {
		panic(err)
	}
	var out [32]byte
	copy(out[:], h.Sum(nil))
	return out
}

func SignSeal(signer Signer, name string, ts time.Time, mr [32]byte) ([65]byte, error) {
	h := SealHash(name, ts, mr)
	return signer.Sign(name, h)
}
