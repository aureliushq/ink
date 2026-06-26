---
title: "Parsing the RESP Protocol"
description: Redis clients and servers communicate using RESP. Here we write a small parser that turns raw bytes into commands we can act on.
tags:
  - go
  - redis
  - networking
status: published
createdAt: 2025-01-13
publishedAt: 2025-01-13
updatedAt: 2025-01-13
series_id: build-your-own-redis-in-go
series_order: 2
---

RESP (REdis Serialization Protocol) is delightfully simple. The first byte of every message tells you its type: `+` for simple strings, `-` for errors, `:` for integers, `$` for bulk strings, and `*` for arrays.

A command like `SET name ink` arrives on the wire as an array of bulk strings:

```
*3\r\n$3\r\nSET\r\n$4\r\nname\r\n$3\r\nink\r\n
```

To parse it we read the `*`, learn there are three elements, then read each `$<len>` bulk string in turn. Once we've decoded the array into `["SET", "name", "ink"]` we have a command our server can dispatch on. We'll build that dispatch table next.
