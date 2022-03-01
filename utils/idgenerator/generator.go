package idgenerator

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"log"
	"unsafe"

	"github.com/yu31/snowflake"
)

// IDGenerator implements an ID Generator uses to generate unique ID
// ID format: "prefix" | 16 bytes string
type IDGenerator struct {
	prefix string
	worker *snowflake.Snowflake
}

// New return an new IDGenerator
func New(prefix string, opts ...Option) *IDGenerator {
	cfg := applyOptions(opts...)
	worker, err := snowflake.New(*cfg.instanceId)
	if err != nil {
		panic(fmt.Errorf("IDGenerator: unexpected error %v", err))
	}

	g := &IDGenerator{
		prefix: prefix,
		worker: worker,
	}
	return g
}

// Take return a new unique id that format with `prefix` + `16 bytes string`.
func (g *IDGenerator) Take() (string, error) {
	id, err := g.worker.Next()
	if err != nil {
		log.Printf("IDGenerator: take new id from worker error: %v\n", err)
		return "", err
	}
	return g.encode(id)
}

func (g *IDGenerator) encode(x int64) (string, error) {
	buf := make([]byte, 8)

	binary.BigEndian.PutUint64(buf, uint64(x))

	lp := len(g.prefix)
	dst := make([]byte, lp+hex.EncodedLen(len(buf)))

	copy(dst[:lp], g.prefix)
	hex.Encode(dst[lp:], buf)

	return *(*string)(unsafe.Pointer(&dst)), nil
}
