package core

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcec/v2/schnorr"
)

type Kind uint32

const (
	KindSetMetadata Kind = 0
	KindTextNote    Kind = 1
)

type Event struct {
	Id        string   `json:"id"`
	PubKey    string    `json:"pubkey"`
	CreatedAt Timestamp `json:"created_at"`
	Kind      uint32    `json:"kind"`
	Tags      Tags      `json:"tags"`
	Content   string    `json:"content"`
	Sig       string    `json:"sig"`
}

// To obtain the event id, we sha256 the serialized event.
func (s Event) GetId() string {
	h := sha256.Sum256(s.Serialize())
	return hex.EncodeToString(h[:])
}

func (s Event) String() string {
	bytes, err := json.Marshal(s)
	if err != nil {
		log.Fatalln("Unable to convert event to string")
	}
	return string(bytes)
}

// The serialization is done over the UTF-8 JSON-serialized string (with no white space or line breaks).
// [
//
//	0,
//	<pubkey, as a (lowercase) hex string>,
//	<created_at, as a number>,
//	<kind, as a number>,
//	<tags, as an array of arrays of non-null strings>,
//	<content, as a string>
//
// ]
func (s Event) Serialize() []byte {

	out := make([]byte, 0)

	out = append(out, []byte(
		fmt.Sprintf(
			"[0,\"%s\",%d,%d,",
			s.PubKey,
			s.CreatedAt,
			s.Kind,
		))...)

    // Add encoded tags.
	out = s.Tags.Encode(out)
	out = append(out, ',')

    // Add encoded user content.
	out = append(out, []byte(s.Content)...)
	out = append(out, ']')

	return out
}

func (s *Event) Sign(key string) error {

	log.Printf("signing event with key: %s", key)

	bytes, err := hex.DecodeString(key)
	if err != nil {
		log.Fatalf("unable to decode secret: %v", err)
		return fmt.Errorf("Sign called with invalid private key '%s': %w", key, err)
	}

	if s.Tags == nil {
		s.Tags = make(Tags, 0)
	}

	sk, pk := btcec.PrivKeyFromBytes(bytes)
	pkBytes := pk.SerializeCompressed()
	s.PubKey = hex.EncodeToString(pkBytes[1:])

	h := sha256.Sum256(s.Serialize())
	sig, err := schnorr.Sign(sk, h[:])
	if err != nil {
		return err
	}

	s.Id = hex.EncodeToString(h[:])
	s.Sig = hex.EncodeToString(sig.Serialize())

	log.Printf("event signed with ID: %s", s.Id)

	return nil
}
