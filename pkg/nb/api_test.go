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

func TestUpdateTierParamsMarshal(t *testing.T) {
	params := UpdateTierParams{
		Name:          "tier1",
		DataPlacement: "SPREAD",
		AttachedPools: []string{"pool-a", "pool-b"},
	}
	data, err := json.Marshal(params)
	if err != nil {
		t.Fatal(err)
	}
	expected := `{"name":"tier1","data_placement":"SPREAD","attached_pools":["pool-a","pool-b"]}`
	if string(data) != expected {
		t.Fatalf("unexpected marshal result: %s", string(data))
	}
}

func TestUpdateTierParamsMarshalPoolsOnly(t *testing.T) {
	params := UpdateTierParams{
		Name:          "tier1",
		AttachedPools: []string{"new-pool"},
	}
	data, err := json.Marshal(params)
	if err != nil {
		t.Fatal(err)
	}
	expected := `{"name":"tier1","attached_pools":["new-pool"]}`
	if string(data) != expected {
		t.Fatalf("unexpected marshal result: %s", string(data))
	}
}

func TestTierInfoUnmarshal(t *testing.T) {
	tierJSON := `{"name":"tier1","data_placement":"SPREAD","attached_pools":["pool-a","pool-b"]}`
	tier := TierInfo{}
	err := json.Unmarshal([]byte(tierJSON), &tier)
	if err != nil {
		t.Fatal(err)
	}
	if tier.Name != "tier1" {
		t.Fatalf("expected name=tier1, got %s", tier.Name)
	}
	if tier.DataPlacement != "SPREAD" {
		t.Fatalf("expected data_placement=SPREAD, got %s", tier.DataPlacement)
	}
	if len(tier.AttachedPools) != 2 {
		t.Fatalf("expected 2 attached_pools, got %d", len(tier.AttachedPools))
	}
}
