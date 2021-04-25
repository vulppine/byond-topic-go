package byondtopic

import (
	"bytes"
	"fmt"
	"io"
	"testing"
)

func TestTopicCrafting(t *testing.T) {
	c := NewTopic()
	c.Write([]byte("test"))
	c.Close()

	r, err := io.ReadAll(c)
	if err != nil {
		t.Error(err)
	}

	t.Log(r)
}

func TestTopicReading(t *testing.T) {
	s := []byte{0x00, 0x83, 0x00, 0x01, 0x06, byte('0'), 0x00}

	b := new(bytes.Buffer)
	b.Write(s)

	r, err := readTopic(b)
	if err != nil {
		t.Error(err)
	}

	if r != "0" {
		t.Error(fmt.Errorf("readTopic: incorrect response: expected string 0"))
	}
}
