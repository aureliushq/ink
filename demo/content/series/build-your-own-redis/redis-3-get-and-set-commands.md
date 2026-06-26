---
title: "GET and SET"
description: With parsing in place, we implement an in-memory key/value store and wire up the GET and SET commands so our server actually does something useful.
tags:
  - go
  - redis
status: published
createdAt: 2025-01-20
publishedAt: 2025-01-20
updatedAt: 2025-01-20
series_id: build-your-own-redis-in-go
series_order: 3
---

Now for the payoff. A real Redis does a lot, but the heart of it is a map. We guard it with a mutex so concurrent clients don't corrupt it:

```go
type Store struct {
	mu   sync.RWMutex
	data map[string]string
}

func (s *Store) Set(k, v string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[k] = v
}

func (s *Store) Get(k string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	v, ok := s.data[k]
	return v, ok
}
```

Dispatch the parsed command: `SET` writes to the store and replies `+OK\r\n`; `GET` reads from it and replies with a bulk string (or `$-1\r\n` when the key is missing). That's a working — if minimal — Redis. From here you can add `DEL`, `EXPIRE`, and persistence.
