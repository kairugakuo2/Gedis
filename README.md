# Gedis

A minimal, in-mem database written in Go, speaking the Redis protocol.

Built from scratch as a way to understand databases and what goes on underneath the API (how data is parsed, held in memory, written to disk)

## Status
Learning project and not production software.

## Features

- **RESP protocol** -> parses and serializes the Redis Serialization Protocol, so any client can talk to it
- **Concurrent connections** -> one go-routine per client
- **Commands** -> `PING`, `SET`, `GET`, `HSET`, `HGET`, `HGETALL`
- **Persistence** -> append-only file written on every write and replayed on startup

## License

MIT