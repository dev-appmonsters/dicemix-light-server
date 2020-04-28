package utils

import (
	"testing"
)

type uint64TestPair struct {
	data [][]uint64
	res  bool
}

type bytesTestPair struct {
	operand1 [][]byte
	operand2 [][]byte
	res      bool
}

type byteStringPair struct {
	data []byte
	res  string
}

type stringUintPair struct {
	data string
	res  uint64
}

var subsetTests = []uint64TestPair{
	{
		[][]uint64{
			{1859546079985200847, 1945157946220288822, 2071666930927106951, 1683255082316998317},
			{338987782431557515, 760646884788788847, 805715802280412061, 855681932209541597, 1404356687488594778},
		}, false,
	},
	{
		[][]uint64{
			{775687816567226590, 1737237369846472686, 1451264941915419093, 1707673020921186118},
			{775687816567226590, 1737237369846472686, 370018584425987655, 1458106296827716655, 1451264941915419093, 1707673020921186118},
		}, true,
	},
	{
		[][]uint64{
			{1890557469049449562, 1784695988170892157, 2043447899551859310, 1647265920402198959, 1817920860757799870},
			{1890557469049449562, 1784695988170892157, 2043447899551859310, 1647265920402198959, 1817920860757799870},
		}, true,
	},
}

var checkEqualUint64Tests = []uint64TestPair{
	{
		[][]uint64{
			{1859546079985200847, 1945157946220288822, 2071666930927106951, 1683255082316998317},
			{1859546079985200847, 1945157946220288822, 2071666930927106951, 1683255082316998317},
		}, true,
	},
	{
		[][]uint64{
			{775687816567226590, 1737237369846472686, 1451264941915419093, 1707673020921186118},
			{775687816567226590, 1737237369846472686, 1451264941915419093, 1707673020921186118},
		}, true,
	},
	{
		[][]uint64{
			{1890557469049449562, 1784695988170892157, 1647265920402198959, 1817920860757799870},
			{1890557469049449562, 1784695988170892157, 2043447899551859310, 1647265920402198959, 1817920860757799870},
		}, false,
	},
}

var equalBytesTests = []bytesTestPair{
	{
		[][]byte{
			{74, 52, 42, 106, 218, 90, 24, 152, 112, 204, 87, 191, 57, 232, 101, 218, 78, 71, 83, 126},
			{243, 5, 162, 57, 245, 33, 192, 154, 112, 26, 31, 72, 3, 0, 91, 18, 249, 99, 238, 10},
			{175, 79, 31, 47, 75, 213, 73, 67, 144, 101, 97, 156, 52, 229, 39, 36, 198, 206, 52, 96},
		}, [][]byte{
			{74, 52, 42, 106, 218, 90, 24, 152, 112, 204, 87, 191, 57, 232, 101, 218, 78, 71, 83, 126},
			{243, 5, 162, 57, 245, 33, 192, 154, 112, 26, 31, 72, 3, 0, 91, 18, 249, 99, 238, 10},
			{175, 79, 31, 47, 75, 213, 73, 67, 144, 101, 97, 156, 52, 229, 39, 36, 198, 206, 52, 96},
		}, true,
	},
	{
		[][]byte{
			{5, 162, 57, 245, 33, 192, 154, 112, 26, 31, 72, 3, 0, 91, 18, 249, 99, 238, 10},
			{175, 79, 31, 47, 75, 213, 73, 67, 144, 101, 97, 156, 52, 229, 39, 36, 198, 206, 52, 96},
		}, [][]byte{
			{74, 52, 42, 106, 218, 90, 24, 152, 112, 204, 87, 191, 57, 232, 101, 218, 78, 71, 83, 126},
			{175, 79, 31, 47, 75, 213, 73, 67, 144, 101, 97, 156, 52, 229, 39, 36, 198, 206, 52, 96},
		}, false,
	},
	{
		[][]byte{
			{74, 52, 42, 106, 218, 90, 24, 152, 112, 204, 87, 191, 57, 232, 101, 218, 78, 71, 83, 126},
		}, [][]byte{
			{74, 52, 42, 106, 218, 90, 24, 152, 112, 204, 87, 191, 57, 232, 101, 218, 78, 71, 83, 126},
		}, true,
	},
}

var removeEmptyBytesTests = []bytesTestPair{
	{
		[][]byte{
			{74, 52, 42, 106, 218, 90, 24, 152, 112, 204, 87, 191, 57, 232, 101, 218, 78, 71, 83, 126},
			{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			{175, 79, 31, 47, 75, 213, 73, 67, 144, 101, 97, 156, 52, 229, 39, 36, 198, 206, 52, 96},
		}, [][]byte{
			{74, 52, 42, 106, 218, 90, 24, 152, 112, 204, 87, 191, 57, 232, 101, 218, 78, 71, 83, 126},
			{175, 79, 31, 47, 75, 213, 73, 67, 144, 101, 97, 156, 52, 229, 39, 36, 198, 206, 52, 96},
		}, true,
	},
	{
		[][]byte{
			{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			{175, 79, 31, 47, 75, 213, 73, 67, 144, 101, 97, 156, 52, 229, 39, 36, 198, 206, 52, 96},
		}, [][]byte{
			{175, 79, 31, 47, 75, 213, 73, 67, 144, 101, 97, 156, 52, 229, 39, 36, 198, 206, 52, 96},
		}, true,
	},
}

var bytesToBase58StringTests = []byteStringPair{
	{
		[]byte{191, 196, 191, 88, 228, 52, 126, 202, 91, 18, 50, 157, 255, 25, 125, 91, 8, 75, 201, 85},
		"3fxTZkRBKNj3KmJVkM3du4BP15e4",
	},
	{
		[]byte{47, 65, 205, 229, 8, 1, 35, 24, 239, 112, 76, 187, 142, 253, 154, 143, 217, 141, 150, 142},
		"fBkugJ6GtKuA5ez1Xt29hxfMjfj",
	},
	{
		[]byte{87, 228, 213, 202, 230, 147, 141, 171, 155, 118, 147, 79, 17, 202, 203, 130, 238, 42, 42, 119},
		"2E2FWkeU8Ku4trPYfrCPLxaU79DY",
	},
}

var shortHashTests = []stringUintPair{
	{"2cvc1KdBHVYPH9dhXn5eEGpKpwdt", 1996933178227623167},
	{"mmgHJHSbJM2JycX6zZjdeXNw8sf", 13084505176271441661},
	{"2vyzGBqpScgDMGXsg9HqH4fyyyqq", 8404735959948483678},
}

func TestIsSubset(t *testing.T) {
	for _, pair := range subsetTests {
		output := IsSubset(pair.data[0], pair.data[1])

		if output != pair.res {
			t.Error(
				"For", pair.data,
				"expected", pair.res,
				"got", output,
			)
		}
	}
}

func TestCheckEqualUint64(t *testing.T) {
	for _, pair := range checkEqualUint64Tests {
		output := CheckEqualUint64(pair.data[0], pair.data[1])

		if output != pair.res {
			t.Error(
				"For", pair.data,
				"expected", pair.res,
				"got", output,
			)
		}
	}
}

func TestEqualBytes(t *testing.T) {
	for _, pair := range equalBytesTests {
		output := EqualBytes(pair.operand1, pair.operand2)

		if output != pair.res {
			t.Error(
				"expected", pair.res,
				"got", output,
			)
		}
	}
}

func TestRemoveEmptyBytes(t *testing.T) {
	for _, pair := range removeEmptyBytesTests {
		output := RemoveEmptyBytes(pair.operand1)
		res := EqualBytes(output, pair.operand2)

		if res != pair.res {
			t.Error(
				"expected", pair.res,
				"got", res,
			)
		}
	}
}

func TestBytesToBase58String(t *testing.T) {
	for _, pair := range bytesToBase58StringTests {
		output := BytesToBase58String(pair.data)

		if output != pair.res {
			t.Error(
				"For", pair.data,
				"expected", pair.res,
				"got", output,
			)
		}
	}
}

func TestShortHash(t *testing.T) {
	for _, pair := range shortHashTests {
		output := ShortHash(pair.data)

		if output != pair.res {
			t.Error(
				"For", pair.data,
				"expected", pair.res,
				"got", output,
			)
		}
	}
}
