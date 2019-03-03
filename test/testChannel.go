package main

import (
	"fmt"
	"sync"
	"time"
)

func main() {
	var wg = &sync.WaitGroup{}
	from := makeFrom(wg)
	to := makeOut(wg)


	go func() {
		for i := range from {
			to <- i
		}
	}()

	wg.Wait()
}

func makeOut(wg *sync.WaitGroup) chan int {
	out := make(chan int)
	go func() {
		for i := range out {
			fmt.Println("outer recevied: ",i)
			wg.Done()
		}
	}()
	return out
}

func makeFrom(wg *sync.WaitGroup) chan int {
	out := make(chan int)
	for i:=0;i<10;i++ {
		wg.Add(1)
		go func() {
			time.Sleep(time.Second)
			out <- i
		}()
	}
	return out
}


