package wal

import "hash/crc32"

func checksum(b []byte) uint32 {
	return crc32.ChecksumIEEE(b)
}
