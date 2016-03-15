package main

import (
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/SoheilSalehian/pied-piper/lib"
)

var wg sync.WaitGroup

func main() {
	t0 := time.Now()
	wg.Add(1)
	pipeline := piedpiper.NewPipeline(piedpiper.DownloadAll{}, piedpiper.ConvertToPng{}, piedpiper.CallOcrClient{}, piedpiper.DeletePDF{})
	go func() {
		fmt.Println("inside generation of ids")
		// TODO: Implment actual id generation using API/DB queries
		for i := 2278123; i <= 2278223; i++ {
			ad := piedpiper.NewAd(strconv.Itoa((i)))
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
