package pulse

import (
	"fmt"
	"os"
	"sync/atomic"
	"testing"
	"time"
)

func Benchmark(b *testing.B) {
	const size = 4096 * 1
	var packet [size]byte
	for i := 0; i < size; i++ {
		packet[i] = byte(i % 255)
	}

	for _, parallelism := range []int{1, 4, 16, 32, 64, 128, 256} {
		var (
			routines int32
			writes   int
		)

		started := time.Now()

		b.Run(fmt.Sprintf("Parallel:%d", parallelism), func(b *testing.B) {
			b.SetParallelism(parallelism)
			b.SetBytes(int64(size))

			writes += b.N

			writer, err := Open(
				"test.log",
				os.O_CREATE|os.O_TRUNC,
				0666,
				time.Microsecond*100,
			)
			if err != nil {
				b.Fatalf("unable to open file: %s", err)
			}

			b.ResetTimer()

			b.RunParallel(func(pb *testing.PB) {
				atomic.AddInt32(&routines, 1)

				for pb.Next() {
					writer.Write(packet[:])
				}
			})
		})

		finished := time.Now()

		fmt.Printf(
			"> total concurrency: %d, %.2f writes/routine, %.2f writes/sec\n",
			routines,
			float64(writes)/float64(routines),
			float64(writes)/finished.Sub(started).Seconds(),
		)
	}
}

//func Test(t *testing.T) {
//    test := assert.New(t)
//    test.NoError(err)

//    wg := sync.WaitGroup{}

//    concurrency := 256
//    amount := 10000

//    started := time.Now()
//    for i := 0; i < concurrency; i++ {
//        wg.Add(1)
//        go func() {
//            for n := 0; n < amount/concurrency; n++ {
//                writer.Write(packet[:])
//            }
//            wg.Done()
//        }()
//    }

//    wg.Wait()
//    finished := time.Now()

//    duration := finished.Sub(started)

//    fmt.Printf("%d - %s (%.2f)\n", amount, duration, float64(amount)/duration.Seconds())
//}
