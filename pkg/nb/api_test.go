package nb

import (
	"encoding/json"
	"fmt"
	"testing"
)

func testBigIntMarshal(t *testing.T, bi BigInt) {
	data, err := json.Marshal(bi)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("%+v => %q \n", bi, string(data))
}

func testBigIntUnmarshal(t *testing.T, bigIntJSON string) {
	bi := BigInt{}
	err := json.Unmarshal([]byte(bigIntJSON), &bi)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("Unmarshal %q => %#v \n", bigIntJSON, bi)
}

func TestBigIntMarshalSmall(t *testing.T) {
	testBigIntMarshal(t, BigInt{N: 11})
}

func TestBigIntMarshalBig(t *testing.T) {
	testBigIntMarshal(t, BigInt{N: 666, Peta: 1})
}

func TestBigIntUnmarshalZero(t *testing.T) {
	testBigIntUnmarshal(t, `0`)
}

func TestBigIntUnmarshalSmall(t *testing.T) {
	testBigIntUnmarshal(t, `3333`)
}

func TestBigIntUnmarshalBig(t *testing.T) {
	testBigIntUnmarshal(t, `{"n":1111,"peta":29}`)
}

func TestBigIntUnmarshalBigNoPeta(t *testing.T) {
	testBigIntUnmarshal(t, `{"n":99}`)
}
