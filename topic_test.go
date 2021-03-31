package byondtopic

import (
	"io"
	"testing"
)

func TestTopicCrafting(t *testing.T) {
	c := NewTopic()
	c.Write([]byte("test"))
	c.Close()

	r, err := io.ReadAll(c)
	if err != nil { t.Error(err) }

	t.Log(r)
}
