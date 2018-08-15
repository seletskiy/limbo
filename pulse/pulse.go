package pulse

import (
	"os"
	"sync"
	"time"
)

type Pulse struct {
	lock sync.Mutex

	latency time.Duration

	file   *os.File
	buffer []byte

	writers []chan error
}

func Open(
	name string,
	flag int,
	permission os.FileMode,
	latency time.Duration,
) (*Pulse, error) {
	file, err := os.OpenFile(name, flag|os.O_SYNC|os.O_WRONLY, permission)
	if err != nil {
		return nil, err
	}

	return &Pulse{
		file:    file,
		latency: latency,
	}, nil
}

func (pulse *Pulse) Write(data []byte) (int, error) {
	pulse.lock.Lock()

	errs := make(chan error)

	pulse.buffer = append(pulse.buffer, data...)
	pulse.writers = append(pulse.writers, errs)

	if len(pulse.writers) == 1 {
		time.AfterFunc(pulse.latency, pulse.commit)
	}

	pulse.lock.Unlock()

	err := <-errs
	if err != nil {
		return 0, err
	}

	return len(data), nil
}

func (pulse *Pulse) commit() {
	pulse.lock.Lock()

	//s := time.Now()
	_, err := pulse.file.Write(pulse.buffer)
	//e := time.Now()
	//fmt.Fprintf(os.Stderr, "XXXXXX pulse.go:64 e.Sub(s): %#v\n", e.Sub(s).Seconds())
	if err != nil {
		for _, errs := range pulse.writers {
			errs <- err
		}
	}

	for _, errs := range pulse.writers {
		close(errs)
	}

	pulse.buffer = pulse.buffer[:0]
	pulse.writers = pulse.writers[:0]

	pulse.lock.Unlock()
}
