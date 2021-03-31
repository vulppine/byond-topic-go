package byondtopic

import (
	"bytes"
	// "encoding/binary"
	"fmt"
	"errors"
	"io"
	"net"
)

const (
	BYONDMagicNumber = "\x00\x83"
)

var (
	InvalidTopic = errors.New("recieved message is not valid")
)

type Topic struct {
	buf *bytes.Buffer
	raw []byte
	c bool
}

func NewTopic() *Topic {
	t := new(Topic)
	t.buf = new(bytes.Buffer)

	return t
}

func (t *Topic) Read(p []byte) (int, error) {
	if !t.c {
		return 0, fmt.Errorf("topic was not closed before read")
	}

	return t.buf.Read(p)
}

func (t *Topic) Write(p []byte) (int, error) {
	t.raw = append(t.raw, p...)

	return len(p), nil
}

func (t *Topic) Close() error {
	t.buf.WriteString(BYONDMagicNumber)

	l := uint16(len(t.raw) + 6)

	// shift the highest 8 bits, mask, then write
	_, err := t.buf.Write([]byte{byte((l >> 8) & 255)})
	if err != nil { return err }

	// mask the highest 8 bits, write it
	_, err = t.buf.Write([]byte{byte(l & 255)})
	if err != nil { return err }

	t.buf.Write([]byte{0x00, 0x00, 0x00, 0x00, 0x00})
	t.buf.Write(t.raw)
	t.buf.Write([]byte{0x00})
	t.c = true

	return nil
}

// internal use, there's no real need to export this
// function
func readTopic(r io.Reader) (string, error) {
	head := make([]byte, 5)
	_, err := r.Read(head)

	if err != nil { return "", err }

	// if it doesn't have the correct magic number,
	// or isn't an ASCII string,
	// return InvalidTopic
	if head[0] != 0x00 || head[1] != 0x83 || head[4] != 0x06 {
		return "", InvalidTopic
	}

	var l uint16
	if head[2] == 0x00 {
		// it'll just be the fourth byte as a uint16
		l = uint16(head[3])
	} else {
		// convert the first half to a uint16, shift it left
		l = uint16(head[2]) << 0xFF

		// convert the second half, perform binary AND
		l = l & uint16(head[3])
	}

	m := make([]byte, l)

	_, err = r.Read(m)
	if err != nil { return "", err }

	return string(m), nil
}

// SendTopic sends a named topic to the given address addr.
// addr is assumed to be a BYOND server, and will respond
// accordingly to a BYOND Topic result.
func SendTopic(addr, s string) (string, error) {
	a, _ := net.ResolveTCPAddr("tcp", addr)
	c, err := net.DialTCP("tcp", nil, a)

	defer c.Close()
	if err != nil { return "", err }

	t := NewTopic()
	t.Write([]byte("?" + s))
	t.Close()

	b, err := io.ReadAll(t)
	if err != nil { return "", err }

	go c.Write(b)

	r, err := readTopic(c)
	if err != nil { return "", err }

	return r, nil
}
