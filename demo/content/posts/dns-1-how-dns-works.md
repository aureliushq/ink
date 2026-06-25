---
title: "Build Your Own DNS Server in Go — Part 1: How DNS Works"
description: Before writing any code we map out the DNS request/response flow and the structure of a DNS message so the rest of the series makes sense.
tags:
  - go
  - dns
  - networking
status: published
createdAt: 2025-02-03
publishedAt: 2025-02-03
updatedAt: 2025-02-03
series_id: build-your-own-dns-server-in-go
series_order: 1
---

When you type a domain into your browser, a DNS resolver translates it into an IP address. That conversation happens over UDP on port `53` using a compact binary message format.

Every DNS message has the same shape:

```
+---------------------+
|       Header        |  12 bytes: ID, flags, counts
+---------------------+
|      Question       |  what are we asking about?
+---------------------+
|       Answer        |  resource records
+---------------------+
|     Authority       |
+---------------------+
|     Additional      |
+---------------------+
```

Over this series we'll build a server that reads these messages, understands the question, and sends back an answer. First up: opening a UDP socket.
