---
title: "Parsing the Header"
description: The 12-byte DNS header holds the message ID, flags, and section counts. We decode it into a struct and prepare to answer the question.
tags:
  - go
  - dns
status: published
createdAt: 2025-02-17
publishedAt: 2025-02-17
updatedAt: 2025-02-17
series_id: build-your-own-dns-server-in-go
series_order: 3
---

Every DNS message starts with a fixed 12-byte header. It's a tight little structure packed into big-endian integers:

```go
type Header struct {
	ID      uint16
	Flags   uint16
	QDCount uint16 // questions
	ANCount uint16 // answers
	NSCount uint16 // authority records
	ARCount uint16 // additional records
}

func parseHeader(b []byte) Header {
	return Header{
		ID:      binary.BigEndian.Uint16(b[0:2]),
		Flags:   binary.BigEndian.Uint16(b[2:4]),
		QDCount: binary.BigEndian.Uint16(b[4:6]),
		ANCount: binary.BigEndian.Uint16(b[6:8]),
		NSCount: binary.BigEndian.Uint16(b[8:10]),
		ARCount: binary.BigEndian.Uint16(b[10:12]),
	}
}
```

The `ID` lets a client match a response to its query, and the `Flags` field encodes whether this is a query or response, the opcode, and the response code. To reply, we echo the `ID`, flip the response bit in `Flags`, copy the question, and append an answer record. That's the whole loop — and the foundation for a real resolver.
