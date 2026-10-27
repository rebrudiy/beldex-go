// Package walletrpc is a minimal client for beldex-wallet-rpc (JSON-RPC 2.0).
//
// Field names follow src/wallet/wallet_rpc_server_commands_defs.h in
// Beldex-Coin/beldex. Amounts are in atomic units: 1 BDX = 1e9.
package walletrpc

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync/atomic"
	"time"
)

// AtomicPerBDX is the number of atomic units in one BDX (9 decimals).
const AtomicPerBDX = 1_000_000_000

// PriorityFlash makes a transfer go through the Flash quorum (seconds instead of blocks).
const PriorityFlash = 5

type Client struct {
	url  string
	http *http.Client
	id   atomic.Uint64
}

// New returns a client for a wallet-rpc listening at baseURL, e.g. "http://127.0.0.1:29092".
func New(baseURL string) *Client {
	return &Client{
		url:  baseURL + "/json_rpc",
		http: &http.Client{Timeout: 60 * time.Second},
	}
}

// RPCError is an error returned by the wallet itself (as opposed to a transport error).
type RPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (e *RPCError) Error() string { return fmt.Sprintf("wallet-rpc error %d: %s", e.Code, e.Message) }

// Call invokes method with params and decodes the result into out (out may be nil).
func (c *Client) Call(ctx context.Context, method string, params, out any) error {
	body, err := json.Marshal(map[string]any{
		"jsonrpc": "2.0",
		"id":      c.id.Add(1),
		"method":  method,
		"params":  params,
	})
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("%s: %w", method, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return fmt.Errorf("%s: http %d: %s", method, resp.StatusCode, b)
	}

	var envelope struct {
		Result json.RawMessage `json:"result"`
		Error  *RPCError       `json:"error"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&envelope); err != nil {
		return fmt.Errorf("%s: decode: %w", method, err)
	}
	if envelope.Error != nil {
		return fmt.Errorf("%s: %w", method, envelope.Error)
	}
	if out == nil {
		return nil
	}
	return json.Unmarshal(envelope.Result, out)
}
