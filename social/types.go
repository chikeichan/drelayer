package social

import (
	"encoding/hex"
	"github.com/ddrp-org/dformats"
	"time"
)

type EnvelopeFormatter interface {
	EnvelopeFormat() *dformats.Envelope
}

type Envelope struct {
	Username  string
	TLD       string
	CreatedAt time.Time
	GUID      string
	Refhash   string
}

type Post struct {
	Envelope
	ID        int
	Body      string
	Title     *string
	Reference *string
	Topic     *string
	Tags      []string
}

func (p *Post) EnvelopeFormat() *dformats.Envelope {
	msg := dformats.NewPost(dformats.DefaultVersion, dformats.DefaultSubtype)
	msg.Body = p.Body
	if p.Title != nil {
		msg.Title = *p.Title
	}
	if p.Reference != nil {
		msg.Reference = mustConvertHash(*p.Reference)
	}
	if p.Topic != nil {
		msg.Topic = *p.Topic
	}
	msg.Tags = p.Tags
	return &dformats.Envelope{
		Timestamp: p.CreatedAt,
		GUID:      mustConvertGUID(p.GUID),
		Message:   msg,
	}
}

type Connection struct {
	Envelope
	ID                 int
	ConnecteeTLD       string
	ConnecteeSubdomain *string
	Type               int
}

func (c *Connection) EnvelopeFormat() *dformats.Envelope {
	var msg *dformats.Connection
	switch c.Type {
	case ConnectionTypeFollow:
		msg = dformats.NewFollow()
	case ConnectionTypeBlock:
		msg = dformats.NewBlock()
	default:
		panic("invalid connection type")
	}
	msg.TLD = c.ConnecteeTLD
	if c.ConnecteeSubdomain != nil {
		msg.Subdomain = *c.ConnecteeSubdomain
	}
	return &dformats.Envelope{
		Timestamp: c.CreatedAt,
		GUID:      mustConvertGUID(c.GUID),
		Message:   msg,
	}
}

type Moderation struct {
	Envelope
	ID        int
	Reference string
	Type      int
}

func (m *Moderation) EnvelopeFormat() *dformats.Envelope {
	var msg *dformats.Moderation
	switch m.Type {
	case ModerationTypeLike:
		msg = dformats.NewLike()
	case ModerationTypePin:
		msg = dformats.NewPin()
	default:
		panic("invalid moderation type")
	}
	msg.Reference = mustConvertHash(m.Reference)
	return &dformats.Envelope{
		Timestamp: m.CreatedAt,
		GUID:      mustConvertGUID(m.GUID),
		Message:   msg,
	}
}

func mustConvertGUID(in string) [8]byte {
	if len(in) != 16 {
		panic("invalid guid")
	}
	buf, err := hex.DecodeString(in)
	if err != nil {
		panic(err)
	}
	var out [8]byte
	copy(out[:], buf)
	return out
}

func mustConvertHash(in string) [32]byte {
	if len(in) != 64 {
		panic("invalid hash")
	}
	buf, err := hex.DecodeString(in)
	if err != nil {
		panic(err)
	}
	var out [32]byte
	copy(out[:], buf)
	return out
}
