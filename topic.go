package byondtopic

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
)

var (
	// BYONDMagicNumber is the number that precedes every BYOND topic string.
	BYONDMagicNumber = []byte{0x00, 0x83}
	InvalidTopic     = errors.New("recieved message is not valid")
)

type Topic struct {
	buf *bytes.Buffer
	raw []byte
	c   bool
}

func NewTopic() *Topic {
	t := new(Topic)
	t.buf = new(bytes.Buffer)

	return t
}

// Read reads out a topic into byte slice p. It will always return an error
// if the topic was not closed before it is read.
func (t *Topic) Read(p []byte) (int, error) {
	if !t.c {
		return 0, fmt.Errorf("topic was not closed before read")
	}

	return t.buf.Read(p)
}

// Write writes byte slice p into the topic. It will always return an error
// if the topic is closed beforehand.
func (t *Topic) Write(p []byte) (int, error) {
	if t.c {
		return 0, fmt.Errorf("topic is closed")
	}

	t.raw = append(t.raw, p...)

	return len(p), nil
}

// Close finalizes the topic, and ensures that it is a valid BYOND topic string
// to send. It will return an error if the topic's length is bigger than
// the maximum size of a 16 bit integer.
func (t *Topic) Close() error {
	if len(t.raw) > (1 << 16) {
		return fmt.Errorf("topic is too big")
	}

	t.buf.Write(BYONDMagicNumber)
	b := make([]byte, 2)                                // make our uint16 buffer
	binary.BigEndian.PutUint16(b, uint16(len(t.raw)+6)) // put our topic's length into our buffer (big-endian 8-bit byte pair)
	t.buf.Write(b)                                      // write it

	t.buf.Write([]byte{0x00, 0x00, 0x00, 0x00, 0x00}) // write the spacing
	t.buf.Write(t.raw)                                // write the actual topic
	t.buf.Write([]byte{0x00})                         // write the end of the topic string (null termination)
	t.c = true                                        // declare it to be closed

	return nil
}

// internal use, there's no real need to export this
// function
func readTopic(r io.Reader) (string, error) {
	head := make([]byte, 5)
	_, err := r.Read(head)

	if err != nil {
		return "", err
	}

	// if it doesn't have the correct magic number,
	// or isn't an ASCII string,
	// return InvalidTopic
	if bytes.Compare(BYONDMagicNumber, head[:2]) != 0 || head[4] != 0x06 {
		return "", InvalidTopic
	}

	var l uint16
	if head[2] == 0x00 {
		// it'll just be the fourth byte as a uint16
		l = uint16(head[3])
	} else {
		// use the BigEndian conversion
		l = binary.BigEndian.Uint16(head[2:4])
	}

	m := make([]byte, l)

	_, err = r.Read(m)
	if err != nil {
		return "", err
	}

	return string(m), nil
}

// SendTopic sends a named topic to the given address addr.
// addr is assumed to be a BYOND server, and will respond
// accordingly to a BYOND Topic result.
func SendTopic(addr, s string) (string, error) {
	a, _ := net.ResolveTCPAddr("tcp", addr)
	c, err := net.DialTCP("tcp", nil, a)

	defer c.Close()
	if err != nil {
		return "", err
	}

	t := NewTopic()
	t.Write([]byte("?" + s))
	t.Close()

	b, err := io.ReadAll(t)
	if err != nil {
		return "", err
	}

	go c.Write(b)

	r, err := readTopic(c)
	if err != nil {
		return "", err
	}

	return r, nil
}
