# beldex-go

Minimal Go client for [`beldex-wallet-rpc`](https://github.com/Beldex-Coin/beldex) (JSON-RPC 2.0), no dependencies.

```go
w := walletrpc.New("http://127.0.0.1:29092")

addr, idx, _ := w.NewSubaddress(ctx, "campaign:rent")   // one address per purpose
in, _ := w.Incoming(ctx, idx)                           // confirmed + mempool, Flash-aware
tx, _ := w.Send(ctx, true, walletrpc.Destination{       // true = Flash
    Address: addr, Amount: 5 * walletrpc.AtomicPerBDX,
})
fmt.Println(walletrpc.FormatBDX(tx.Fee))
```

Covers: wallet create/open/restore view-only, subaddresses, view key, incoming transfers
(`flash_mempool`, `was_flash`, `double_spend_seen`), transfers with Flash priority,
`check_tx_key` payment proofs, BDX amount formatting/parsing (9 decimals).

Field names follow `src/wallet/wallet_rpc_server_commands_defs.h` in Beldex-Coin/beldex.

`go run ./cmd/smoke [wallet-rpc-url]` exercises the client against a live wallet.
