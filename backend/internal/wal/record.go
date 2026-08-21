package wal

import (
	"encoding/binary"
	"errors"
	"io"
)

const magic byte = 1

var (
	ErrIncomplete = errors.New("incomplete wal record")
	ErrCorrupt    = errors.New("corrupt wal record")
)

// Record is one WAL entry. Offset is assigned by the log before encode.
type Record struct {
	Offset    int64
	Timestamp int64
	Key       []byte
	Value     []byte
}

// Encode writes:
// crc32(4) size(4) magic(1) offset(8) timestamp(8) keyLen(4) key valueLen(4) value
func Encode(rec Record) []byte {
	keyLen := len(rec.Key)
	valLen := len(rec.Value)
	size := 1 + 8 + 8 + 4 + keyLen + 4 + valLen
	buf := make([]byte, 8+size)
	binary.BigEndian.PutUint32(buf[4:8], uint32(size))
	buf[8] = magic
	binary.BigEndian.PutUint64(buf[9:17], uint64(rec.Offset))
	binary.BigEndian.PutUint64(buf[17:25], uint64(rec.Timestamp))
	binary.BigEndian.PutUint32(buf[25:29], uint32(keyLen))
	copy(buf[29:29+keyLen], rec.Key)
	off := 29 + keyLen
	binary.BigEndian.PutUint32(buf[off:off+4], uint32(valLen))
	copy(buf[off+4:], rec.Value)
	crc := checksum(buf[8:])
	binary.BigEndian.PutUint32(buf[0:4], crc)
	return buf
}

// Decode reads one record. n is bytes consumed on success.
// io.EOF: clean end of file.
// ErrIncomplete: truncated tail (caller should truncate).
// ErrCorrupt: CRC/magic mismatch (caller should truncate from this position).
func Decode(r io.Reader) (rec Record, n int, err error) {
	hdr := make([]byte, 8)
	got, err := io.ReadFull(r, hdr)
	if err == io.EOF && got == 0 {
		return Record{}, 0, io.EOF
	}
	if err != nil {
		return Record{}, got, ErrIncomplete
	}
	crcWant := binary.BigEndian.Uint32(hdr[0:4])
	size := binary.BigEndian.Uint32(hdr[4:8])
	if size < 21 || size > 32*1024*1024 {
		return Record{}, 8, ErrCorrupt
	}
	payload := make([]byte, size)
	got2, err := io.ReadFull(r, payload)
	if err != nil {
		return Record{}, 8 + got2, ErrIncomplete
	}
	if payload[0] != magic {
		return Record{}, 8 + int(size), ErrCorrupt
	}
	if checksum(payload) != crcWant {
		return Record{}, 8 + int(size), ErrCorrupt
	}
	rec.Offset = int64(binary.BigEndian.Uint64(payload[1:9]))
	rec.Timestamp = int64(binary.BigEndian.Uint64(payload[9:17]))
	keyLen := binary.BigEndian.Uint32(payload[17:21])
	if 21+keyLen+4 > size {
		return Record{}, 8 + int(size), ErrCorrupt
	}
	rec.Key = append([]byte(nil), payload[21:21+keyLen]...)
	valOff := 21 + keyLen
	valLen := binary.BigEndian.Uint32(payload[valOff : valOff+4])
	if valOff+4+valLen != size {
		return Record{}, 8 + int(size), ErrCorrupt
	}
	rec.Value = append([]byte(nil), payload[valOff+4:]...)
	return rec, 8 + int(size), nil
}
