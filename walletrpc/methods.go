package walletrpc

import "context"

// ---- wallet lifecycle ----

func (c *Client) CreateWallet(ctx context.Context, filename, password string) error {
	return c.Call(ctx, "create_wallet", map[string]any{
		"filename": filename, "password": password, "language": "English",
	}, nil)
}

func (c *Client) OpenWallet(ctx context.Context, filename, password string) error {
	return c.Call(ctx, "open_wallet", map[string]any{"filename": filename, "password": password}, nil)
}

// GenerateViewOnly restores a view-only wallet: it can see incoming funds but cannot spend.
// This is what makes a campaign or a ballot box publicly auditable.
func (c *Client) GenerateViewOnly(ctx context.Context, filename, address, viewKey string, restoreHeight uint64) error {
	return c.Call(ctx, "generate_from_keys", map[string]any{
		"filename":       filename,
		"address":        address,
		"viewkey":        viewKey,
		"restore_height": restoreHeight,
		"password":       "",
	}, nil)
}

func (c *Client) Refresh(ctx context.Context) error {
	return c.Call(ctx, "refresh", map[string]any{}, nil)
}

// ---- keys and addresses ----

func (c *Client) ViewKey(ctx context.Context) (string, error) {
	var r struct {
		Key string `json:"key"`
	}
	err := c.Call(ctx, "query_key", map[string]any{"key_type": "view_key"}, &r)
	return r.Key, err
}

func (c *Client) PrimaryAddress(ctx context.Context) (string, error) {
	var r struct {
		Address string `json:"address"`
	}
	err := c.Call(ctx, "get_address", map[string]any{"account_index": 0}, &r)
	return r.Address, err
}

// NewSubaddress creates a fresh subaddress in account 0. One per campaign / poll option.
func (c *Client) NewSubaddress(ctx context.Context, label string) (address string, index uint32, err error) {
	var r struct {
		Address      string `json:"address"`
		AddressIndex uint32 `json:"address_index"`
	}
	err = c.Call(ctx, "create_address", map[string]any{"account_index": 0, "label": label}, &r)
	return r.Address, r.AddressIndex, err
}

func (c *Client) Height(ctx context.Context) (uint64, error) {
	var r struct {
		Height uint64 `json:"height"`
	}
	err := c.Call(ctx, "get_height", map[string]any{}, &r)
	return r.Height, err
}

// ---- balances and incoming transfers ----

type SubaddressIndex struct {
	Major uint32 `json:"major"`
	Minor uint32 `json:"minor"`
}

type Transfer struct {
	TxID          string          `json:"txid"`
	Height        uint64          `json:"height"`
	Timestamp     uint64          `json:"timestamp"`
	Amount        uint64          `json:"amount"`
	Type          string          `json:"type"`               // "in", "out", "pool", ...
	TokenID       string          `json:"token_id,omitempty"` // empty = native BDX
	Locked        bool            `json:"locked"`
	SubaddrIndex  SubaddressIndex `json:"subaddr_index"`
	Confirmations uint64          `json:"confirmations"`
	FlashMempool  bool            `json:"flash_mempool"` // approved Flash tx still in the mempool
	WasFlash      bool            `json:"was_flash"`
	DoubleSpend   bool            `json:"double_spend_seen"`
}

// Incoming returns confirmed and mempool incoming transfers for the given subaddress indices
// of account 0 (all subaddresses when indices is empty).
func (c *Client) Incoming(ctx context.Context, indices ...uint32) ([]Transfer, error) {
	params := map[string]any{"in": true, "pool": true, "account_index": 0}
	if len(indices) == 0 {
		params["all_accounts"] = true
	} else {
		params["subaddr_indices"] = indices
	}
	var r struct {
		In   []Transfer `json:"in"`
		Pool []Transfer `json:"pool"`
	}
	if err := c.Call(ctx, "get_transfers", params, &r); err != nil {
		return nil, err
	}
	return append(r.In, r.Pool...), nil
}

// ---- sending ----

type Destination struct {
	Address string `json:"address"`
	Amount  uint64 `json:"amount"`
	TokenID string `json:"token_id,omitempty"` // empty = native BDX
}

type SentTx struct {
	TxHash string `json:"tx_hash"`
	TxKey  string `json:"tx_key"`
	Amount uint64 `json:"amount"`
	Fee    uint64 `json:"fee"`
}

// Send transfers to one or more destinations. flash=true routes it through the Flash quorum.
func (c *Client) Send(ctx context.Context, flash bool, dests ...Destination) (SentTx, error) {
	priority := 1
	if flash {
		priority = PriorityFlash
	}
	var r SentTx
	err := c.Call(ctx, "transfer", map[string]any{
		"destinations": dests,
		"priority":     priority,
		"get_tx_key":   true,
	}, &r)
	return r, err
}

// CheckTxKey proves a payment of a given tx to address using the sender's tx key.
type TxCheck struct {
	Received      uint64          `json:"received"`
	TokenReceived []TokenReceived `json:"token_received,omitempty"`
	InPool        bool            `json:"in_pool"`
	Confirmations uint64          `json:"confirmations"`
}

type TokenReceived struct {
	TokenID string `json:"token_id"`
	Amount  uint64 `json:"amount"`
}

func (c *Client) CheckTxKey(ctx context.Context, txid, txKey, address string) (TxCheck, error) {
	var r TxCheck
	err := c.Call(ctx, "check_tx_key", map[string]any{"txid": txid, "tx_key": txKey, "address": address}, &r)
	return r, err
}
