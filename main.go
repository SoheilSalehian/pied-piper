package main

import (
	"flag"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/SoheilSalehian/pied-piper/lib"
)

var wg sync.WaitGroup

func main() {
	t0 := time.Now()
	var lower = flag.Int("lower", 0, "lower limit")
	var upper = flag.Int("upper", 0, "upper limit")
	flag.Parse()
	wg.Add(1)
	pipeline := piedpiper.NewPipeline(piedpiper.DownloadAll{})
	go func() {
		fmt.Println("inside generation of ids")
		for i := *lower; i <= *upper; i++ {
			ad := piedpiper.NewAd(strconv.Itoa(i))
			pipeline.Enqueue(ad)
		}
		wg.Done()
		pipeline.Close()
	}()

	total := 0
	pipeline.Dequeue(func(a *piedpiper.Ad) {
		total += 1
	})
	fmt.Println("total: ", total)
	wg.Wait()
	t1 := time.Now()
	fmt.Println("Elapsed time: ", t1.Sub(t0))
}
