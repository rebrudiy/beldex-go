package walletrpc

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

// fakeWallet answers JSON-RPC calls from a method -> result map and records params.
func fakeWallet(t *testing.T, results map[string]any, seen map[string]map[string]any) *Client {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/json_rpc" {
			t.Errorf("path = %s, want /json_rpc", r.URL.Path)
		}
		var req struct {
			Method string         `json:"method"`
			Params map[string]any `json:"params"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatal(err)
		}
		if seen != nil {
			seen[req.Method] = req.Params
		}
		res, ok := results[req.Method]
		if !ok {
			json.NewEncoder(w).Encode(map[string]any{"error": map[string]any{"code": -32601, "message": "Method not found"}})
			return
		}
		json.NewEncoder(w).Encode(map[string]any{"result": res})
	}))
	t.Cleanup(srv.Close)
	return New(srv.URL)
}

func TestSendFlashSetsPriority(t *testing.T) {
	seen := map[string]map[string]any{}
	c := fakeWallet(t, map[string]any{
		"transfer": map[string]any{"tx_hash": "abc", "tx_key": "k", "amount": 5 * AtomicPerBDX, "fee": 100},
	}, seen)

	tx, err := c.Send(context.Background(), true, Destination{Address: "bxTest", Amount: 5 * AtomicPerBDX})
	if err != nil {
		t.Fatal(err)
	}
	if tx.TxHash != "abc" || tx.TxKey != "k" {
		t.Fatalf("unexpected tx: %+v", tx)
	}
	if got := seen["transfer"]["priority"]; got != float64(PriorityFlash) {
		t.Fatalf("priority = %v, want %d", got, PriorityFlash)
	}
	if seen["transfer"]["get_tx_key"] != true {
		t.Fatal("get_tx_key not requested")
	}
}

func TestIncomingMergesPool(t *testing.T) {
	c := fakeWallet(t, map[string]any{
		"get_transfers": map[string]any{
			"in":   []any{map[string]any{"txid": "a", "amount": 1, "type": "in", "subaddr_index": map[string]any{"major": 0, "minor": 3}}},
			"pool": []any{map[string]any{"txid": "b", "amount": 2, "type": "pool", "flash_mempool": true}},
		},
	}, nil)

	ts, err := c.Incoming(context.Background(), 3)
	if err != nil {
		t.Fatal(err)
	}
	if len(ts) != 2 || ts[0].SubaddrIndex.Minor != 3 || !ts[1].FlashMempool {
		t.Fatalf("unexpected transfers: %+v", ts)
	}
}

func TestRPCErrorIsTyped(t *testing.T) {
	c := fakeWallet(t, map[string]any{}, nil)
	_, err := c.ViewKey(context.Background())
	var rpcErr *RPCError
	if !errors.As(err, &rpcErr) || rpcErr.Code != -32601 {
		t.Fatalf("err = %v, want RPCError -32601", err)
	}
}
