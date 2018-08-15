Sync file writer that scales with number of concurrent go-routines writing
data.

```
$ go test -run none -bench .
goos: linux
goarch: amd64
pkg: github.com/seletskiy/limbo/pulse
Benchmark/Parallel:1-4              3000            521592 ns/op           7.85 MB/s
> total concurrency: 12, 258.42 writes/routine, 1892.31 writes/sec
Benchmark/Parallel:4-4             10000            102927 ns/op          39.80 MB/s
> total concurrency: 48, 210.44 writes/routine, 9682.23 writes/sec
Benchmark/Parallel:16-4            50000             33588 ns/op         121.95 MB/s
> total concurrency: 256, 234.77 writes/routine, 28996.31 writes/sec
Benchmark/Parallel:32-4           100000             21860 ns/op         187.37 MB/s
> total concurrency: 512, 215.04 writes/routine, 45624.02 writes/sec
Benchmark/Parallel:64-4           100000             14483 ns/op         282.80 MB/s
> total concurrency: 1280, 125.08 writes/routine, 65938.73 writes/sec
Benchmark/Parallel:128-4          100000             13814 ns/op         296.51 MB/s
> total concurrency: 2048, 53.76 writes/routine, 72025.49 writes/sec
Benchmark/Parallel:256-4          100000             11975 ns/op         342.04 MB/s
> total concurrency: 4096, 26.88 writes/routine, 81985.21 writes/sec
PASS
ok      github.com/seletskiy/limbo/pulse        12.470s
```
