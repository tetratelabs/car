# Notable rationale of car

## Why internal/test/httptest?

Go's `net/http/httptest.NewServer` listens on loopback TCP. Network I/O is
[not durably blocking][synctest-blocking], so goroutines stuck on TCP prevent
a [synctest][synctest-pkg] bubble from becoming idle. The fake clock never
advances, and any `time.After` or `time.Sleep` in the code under test hangs.

Our `httptest.NewServer` replaces the TCP listener with `net.Pipe`, following
the pattern in Go's [TestTLSServerWithoutTLSConn][go-serve] and
[TransportCancelRequestBeforeResponseHeaders][go-transport] tests. Pipe
operations block on channels, which are durably blocking, so synctest can
see when the bubble is idle and advance the clock. Retrofitting
`httptest.NewServer` keeps test practice familiar while avoiding the
[blocking][synctest-blocking] behavior.

We also provide `httptest.HTTPClient`, which runs a handler synchronously in
the caller's goroutine with no I/O at all. Tests who need `*http.Client`, but
don't need a real server can use this instead.

## Why Tools.mk?

`Tools.mk` lists Go tools (linters, formatters) with pinned versions so `make`
can run them via `go run package@version` without a platform-specific install.
Tool deps stay out of `go.mod`, which matters because car is imported as a
library.

We chose this over `go tool` (which stores tools in `go.mod`) for two reasons.
First, `go tool` forces all tools into one dependency graph where they can
revlock each other. Second, [Go rejects `-modfile` in workspace
mode][go-work-modfile], forcing `GOWORK=off` into every tool invocation.

`go run` resolves each tool independently, avoiding both problems. The tradeoff
is no `go.sum` lock for tool versions and only working with make.

---
[go-work-modfile]: https://go.dev/issue/59996
[synctest-pkg]: https://pkg.go.dev/testing/synctest
[synctest-blocking]: https://github.com/golang/go/blob/master/src/testing/synctest/synctest.go#L86-L93
[go-serve]: https://github.com/golang/go/blob/master/src/net/http/serve_test.go#L1767
[go-transport]: https://github.com/golang/go/blob/master/src/net/http/transport_test.go#L3079
