package identity

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sync/atomic"
	"time"
)

type Generator interface{ New(prefix string) string }

type Random struct{}

func (Random) New(prefix string) string {
	var b [8]byte
	if _, err := rand.Read(b[:]); err != nil {
		return fmt.Sprintf("%s-%d", prefix, time.Now().UnixNano())
	}
	return prefix + "-" + hex.EncodeToString(b[:])
}

type Sequence struct{ next atomic.Uint64 }

func (s *Sequence) New(prefix string) string {
	return fmt.Sprintf("%s-%06d", prefix, s.next.Add(1))
}
