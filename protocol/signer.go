package protocol

import (
	"encoding/hex"

	"github.com/btcsuite/btcd/btcec"
	"github.com/ddrp-org/ddrp/crypto"
	"github.com/pkg/errors"
)

type SignerFunc func(hash crypto.Hash) ([65]byte, error)

type KeysNames struct {
	Name string
	Key  string
}

type NameSigner struct {
	keys map[string]*btcec.PrivateKey
}

func NewNameSigner(kns []KeysNames) (*NameSigner, error) {
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

func (n *NameSigner) SingleSigner(tld string) crypto.Signer {
	pk := n.keys[tld]
	if pk == nil {
		panic("no private key for name")
	}
	return crypto.NewSECP256k1Signer(n.keys[tld])
}
