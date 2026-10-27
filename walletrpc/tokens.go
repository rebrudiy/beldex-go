package walletrpc

import (
	"context"
	"encoding/json"
)

// Privacy Tokens (HF21+, feature-privacy-token branch). Amounts are in the
// token's own atomic units (see TokenDescriptor.DecimalPoint).

// TokenDescriptor is the JSON registered on-chain for a token. Ticker,
// FullName and TotalMaxSupply are fixed at registration.
type TokenDescriptor struct {
	Ticker         string `json:"ticker"`
	FullName       string `json:"full_name"`
	TotalMaxSupply uint64 `json:"total_max_supply"`
	DecimalPoint   uint8  `json:"decimal_point"`
	MetaInfo       string `json:"meta_info,omitempty"`
}

type RegisteredToken struct {
	TokenID  string `json:"token_id"`
	TxHash   string `json:"tx_hash"`
	Ticker   string `json:"ticker"`
	FullName string `json:"full_name"`
	TxFee    uint64 `json:"tx_fee"`
}

// RegisterToken registers a new token owned by this wallet. Costs a burn fee
// plus a bond locked for ~180 days (1,000 + 10,000 BDX at the time of writing).
func (c *Client) RegisterToken(ctx context.Context, d TokenDescriptor) (RegisteredToken, error) {
	desc, err := json.Marshal(d)
	if err != nil {
		return RegisteredToken{}, err
	}
	var r RegisteredToken
	err = c.Call(ctx, "register_privacy_token", map[string]any{"json_string": string(desc), "get_tx_key": true}, &r)
	return r, err
}

type TokenInfo struct {
	TokenID       string `json:"token_id"`
	FullName      string `json:"full_name"`
	Ticker        string `json:"ticker"`
	MaxSupply     uint64 `json:"max_supply"`
	CurrentSupply uint64 `json:"current_supply"`
}

// OwnedTokens lists tokens registered by this wallet.
func (c *Client) OwnedTokens(ctx context.Context) ([]TokenInfo, error) {
	var r struct {
		Tokens []TokenInfo `json:"tokens"`
	}
	err := c.Call(ctx, "get_owned_tokens", map[string]any{}, &r)
	return r.Tokens, err
}

// MintToken mints amount of an owned token into this wallet.
func (c *Client) MintToken(ctx context.Context, tokenID string, amount uint64) (SentTx, error) {
	var r SentTx
	err := c.Call(ctx, "mint_token", map[string]any{"token_id": tokenID, "amount": amount, "get_tx_key": true}, &r)
	return r, err
}

// SendToken transfers a token to one or more destinations (Destination.TokenID is filled in).
func (c *Client) SendToken(ctx context.Context, tokenID string, dests ...Destination) (SentTx, error) {
	for i := range dests {
		dests[i].TokenID = tokenID
	}
	var r SentTx
	err := c.Call(ctx, "transfer", map[string]any{"destinations": dests, "priority": 1, "get_tx_key": true}, &r)
	return r, err
}
