package signals

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
)

// Context returns a context which will be canceled when either the SIGINT or
// SIGTERM signal is caught. If either signal is caught four more times, the
// program is terminated immediately with exit code 1.
func Context() context.Context {
	ctx, cancel := context.WithCancel(context.Background())
	sigCh := make(chan os.Signal)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigCh
		cancel()
		for i := 0; i < 4; i++ {
			sig = <-sigCh
		}
		log.Fatalf(
			`Received signal "%s" repeatedly; exiting immediately`,
			sig,
		)
	}()
	return ctx
}
