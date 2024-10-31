# dbench

An experimental client package within the littr project repository. Its purpose is to implement new sync/concurrent configurations for the structures with shared access, and to properly evaluate the benchmarks.

Another pkg's purpose is to experiment with the concurrency in general, to build multi-threaded applications and to learn new concepts of the Go language.

## how to run (local Go runtime)

The simpliest (maybe ever) way to run this code is to chenge the directory (`PWD`/`CWD`) to `cmd/dbench/`, and to execute a simple `go run` procedure.

```
cd cmd/dbench
go run ./...
```

## race conditions detector

```
go run -race ./...
```

+ https://go.dev/blog/race-detector

