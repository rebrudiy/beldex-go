package walletrpc

import (
	"fmt"
	"math/big"
	"strings"
)

// FormatBDX renders atomic units as a human BDX amount, trimming trailing zeros: 1500000000 -> "1.5".
func FormatBDX(atomic uint64) string {
	whole, frac := atomic/AtomicPerBDX, atomic%AtomicPerBDX
	if frac == 0 {
		return fmt.Sprintf("%d", whole)
	}
	return strings.TrimRight(fmt.Sprintf("%d.%09d", whole, frac), "0")
}

// ParseBDX parses a decimal BDX amount ("12.5") into atomic units.
func ParseBDX(s string) (uint64, error) {
	r, ok := new(big.Rat).SetString(strings.TrimSpace(s))
	if !ok || r.Sign() < 0 {
		return 0, fmt.Errorf("invalid BDX amount %q", s)
	}
	r.Mul(r, new(big.Rat).SetInt64(AtomicPerBDX))
	if !r.IsInt() {
		return 0, fmt.Errorf("BDX amount %q has more than 9 decimals", s)
	}
	n := r.Num()
	if !n.IsUint64() {
		return 0, fmt.Errorf("BDX amount %q too large", s)
	}
	return n.Uint64(), nil
}
