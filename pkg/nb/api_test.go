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

func TestSafeReplacePoolParamsMarshal(t *testing.T) {
	params := SafeReplacePoolParams{
		OldPoolName:     "old-pool",
		NewPoolName:     "new-pool",
		EnableMigration: true,
	}
	data, err := json.Marshal(params)
	if err != nil {
		t.Fatal(err)
	}
	expected := `{"old_pool_name":"old-pool","new_pool_name":"new-pool","enable_migration":true}`
	if string(data) != expected {
		t.Fatalf("unexpected marshal result: %s", string(data))
	}
}

func TestSafeReplacePoolParamsMarshalNoMigration(t *testing.T) {
	params := SafeReplacePoolParams{
		OldPoolName: "old-pool",
		NewPoolName: "new-pool",
	}
	data, err := json.Marshal(params)
	if err != nil {
		t.Fatal(err)
	}
	// enable_migration should be omitted when false
	expected := `{"old_pool_name":"old-pool","new_pool_name":"new-pool"}`
	if string(data) != expected {
		t.Fatalf("unexpected marshal result: %s", string(data))
	}
}

func TestSafeReplacePoolReplyUnmarshal(t *testing.T) {
	replyJSON := `{"replaced_tiers":3,"updated_accounts":2,"mode":"MIRROR_STARTED"}`
	reply := SafeReplacePoolReply{}
	err := json.Unmarshal([]byte(replyJSON), &reply)
	if err != nil {
		t.Fatal(err)
	}
	if reply.ReplacedTiers != 3 {
		t.Fatalf("expected replaced_tiers=3, got %d", reply.ReplacedTiers)
	}
	if reply.UpdatedAccounts != 2 {
		t.Fatalf("expected updated_accounts=2, got %d", reply.UpdatedAccounts)
	}
	if reply.Mode != "MIRROR_STARTED" {
		t.Fatalf("expected mode=MIRROR_STARTED, got %s", reply.Mode)
	}
}

func TestSafeReplacePoolReplyUnmarshalReplaced(t *testing.T) {
	replyJSON := `{"replaced_tiers":1,"updated_accounts":0,"mode":"REPLACED"}`
	reply := SafeReplacePoolReply{}
	err := json.Unmarshal([]byte(replyJSON), &reply)
	if err != nil {
		t.Fatal(err)
	}
	if reply.ReplacedTiers != 1 {
		t.Fatalf("expected replaced_tiers=1, got %d", reply.ReplacedTiers)
	}
	if reply.UpdatedAccounts != 0 {
		t.Fatalf("expected updated_accounts=0, got %d", reply.UpdatedAccounts)
	}
	if reply.Mode != "REPLACED" {
		t.Fatalf("expected mode=REPLACED, got %s", reply.Mode)
	}
}
