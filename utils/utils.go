package utils

import (
	"bytes"
	"math/rand"
	"time"

	"github.com/dev-appmonsters/dicemix-light-server/field"

	base58 "github.com/jbenet/go-base58"
	"github.com/shomali11/util/xhashes"
)

// IsSubset returns true if the first array is completely
// contained in the second array. There must be at least
// the same number of duplicate values in second as there
// are in first.
func IsSubset(first, second []uint64) bool {
	set := make(map[uint64]int)
	for _, value := range second {
		set[value]++
	}

	for _, value := range first {
		value = Reduce(value)
		if count, found := set[value]; !found {
			return false
		} else if count < 1 {
			return false
		} else {
			set[value] = count - 1
		}
	}
	return true
}

// EqualBytes - compares two slice's of [][]byte for equality
func EqualBytes(slice1, slice2 [][]byte) bool {
	// if slice1 is nil then slice2 should also be nil
	if (slice1 == nil) != (slice2 == nil) {
		return false
	}

	if len(slice1) != len(slice2) {
		return false
	}

	// for each []byte of slice1, slice2
	for i := 0; i < len(slice1); i++ {
		// if slice1[i] != slice2[i] {return false}
		if !bytes.Equal(slice1[i], slice2[i]) {
			return false
		}
	}
	return true
}

// ContainBytes returns true if the first array is completely
// contained in the second array. There must be at least
// the same number of duplicate values in second as there
// are in first.
func ContainBytes(slice1, slice2 [][]byte) bool {
	for _, s1 := range slice1 {
		found := false
		for _, s2 := range slice2 {
			if !bytes.Equal(s1, s2) {
				continue
			}
			found = true
			break
		}
		if !found {
			return false
		}
	}
	return true
}

// CheckEqualUint64 - compares two slice's of []uint64 for equality
func CheckEqualUint64(slice1, slice2 []uint64) bool {
	// if slice1 is nil then slice2 should also be nil
	if (slice1 == nil) != (slice2 == nil) {
		return false
	}

	if len(slice1) != len(slice2) {
		return false
	}

	// for each []byte of slice1, slice2
	for i := 0; i < len(slice1); i++ {
		// if slice1[i] != slice2[i] {return false}
		if slice1[i] != slice2[i] {
			return false
		}
	}
	return true
}

// RemoveEmptyBytes - removes empty byte slices from messages
func RemoveEmptyBytes(messages [][]byte) [][]byte {
	emptyByte := make([]byte, 20)
	output := make([][]byte, 0)

	for _, message := range messages {
		if !bytes.Equal(message, emptyByte) {
			output = append(output, message)
		}
	}
	return output
}

// BytesToBase58String - converts []byte to Base58 Encoded string
func BytesToBase58String(bytes []byte) string {
	return base58.Encode(bytes)
}

// ShortHash - returns 64bit hash of input string message
func ShortHash(message string) uint64 {
	// NOTE: after DC-EXP roots would contain hash reduced into field
	// (as final result would be in field)
	return xhashes.FNV64(message)
}

// Reduce - reduces value into field range
func Reduce(value uint64) uint64 {
	return field.NewField(value).Value()
}

// Timestamp - to identify time of occurence of an event
// returns current timestamp
// example - 2018-08-07 12:04:46.456601867 +0000 UTC m=+0.000753626
func Timestamp() string {
	return time.Now().String()
}

// RandUint64 - returns random uint64
func RandUint64() uint64 {
	rand.Seed(time.Now().UnixNano())
	return rand.Uint64()
}

// RandInt31 -  Int31 returns a non-negative pseudo-random 31-bit integer as an int32
func RandInt31() int32 {
	rand.Seed(time.Now().UnixNano())
	return rand.Int31()
}

// Power parameter sdhould be within uint64 range
func Power(value, t uint64) uint64 {
	return field.NewField(value).Mul(field.NewField(t)).Value()
}
