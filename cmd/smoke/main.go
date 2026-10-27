// smoke runs the walletrpc client against a live wallet-rpc.
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/rebrudiy/beldex-go/walletrpc"
)

func main() {
	url := "http://127.0.0.1:29092"
	if len(os.Args) > 1 {
		url = os.Args[1]
	}
	ctx := context.Background()
	c := walletrpc.New(url)

	if err := c.Refresh(ctx); err != nil {
		log.Fatal(err)
	}
	h, err := c.Height(ctx)
	if err != nil {
		log.Fatal(err)
	}
	addr, idx, err := c.NewSubaddress(ctx, "smoke-campaign")
	if err != nil {
		log.Fatal(err)
	}
	vk, err := c.ViewKey(ctx)
	if err != nil {
		log.Fatal(err)
	}
	in, err := c.Incoming(ctx, idx)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("height=%d\nsubaddress[%d]=%s\nviewkey=%s…\nincoming=%d\n", h, idx, addr, vk[:8], len(in))
}
