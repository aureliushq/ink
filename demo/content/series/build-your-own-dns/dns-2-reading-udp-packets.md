---
title: "Reading UDP Packets"
description: DNS speaks UDP. We bind to port 53 and read raw query packets into a buffer, ready to be decoded.
tags:
  - go
  - dns
  - networking
status: published
createdAt: 2025-02-10
publishedAt: 2025-02-10
updatedAt: 2025-02-10
series_id: build-your-own-dns-server-in-go
series_order: 2
---

Unlike Redis, DNS is connectionless — each query is a single UDP datagram. We listen with `ListenUDP` and read packets into a fixed buffer (512 bytes is the classic DNS message limit):

```go
addr := net.UDPAddr{Port: 53, IP: net.ParseIP("0.0.0.0")}
conn, err := net.ListenUDP("udp", &addr)
if err != nil {
	log.Fatal(err)
}
defer conn.Close()

buf := make([]byte, 512)
for {
	n, client, err := conn.ReadFromUDP(buf)
	if err != nil {
		continue
	}
	go handle(conn, client, buf[:n])
}
```

Each `buf[:n]` is a complete DNS query waiting to be parsed. In the next part we crack open that 12-byte header.
