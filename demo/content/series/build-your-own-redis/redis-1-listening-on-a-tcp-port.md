---
title: "Listening on a TCP Port"
description: Kicking off the series by opening a TCP socket and accepting client connections, the foundation every Redis server is built on.
tags:
  - go
  - redis
  - networking
status: published
createdAt: 2025-01-06
publishedAt: 2025-01-06
updatedAt: 2025-01-06
series_id: build-your-own-redis-in-go
series_order: 1
---

Redis is, at its core, a server that speaks a simple text protocol over TCP. Before we can parse a single command, we need something to listen for connections.

```go
ln, err := net.Listen("tcp", ":6379")
if err != nil {
	log.Fatal(err)
}
defer ln.Close()

for {
	conn, err := ln.Accept()
	if err != nil {
		continue
	}
	go handle(conn)
}
```

That's it for part one. We bind to port `6379` (Redis' default), accept connections in a loop, and hand each one off to a goroutine so multiple clients can talk to us at once. In the next part we'll start reading bytes off the wire.
