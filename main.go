package main

import (
	"context"
	"sync"
)

func producer1() <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)

		for i := 0; i < 10; i++ {
			out <- i
		}
	}()

	return out
}

func producer2() <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)

		for i := 11; i < 20; i++ {
			out <- i
		}
	}()

	return out
}

func consumer(ctx context.Context, in ...<-chan int) <-chan int {
	var wg sync.WaitGroup

	out := make(chan int)

	multiplex := func(c <-chan int) {
		defer wg.Done()
		for i := range c {
			select {
			case <-ctx.Done():
				return
			case out <- i:
			}
		}
	}

	wg.Add(len(in))
	for _, c := range in {
		go multiplex(c)
	}

	go func() {
		wg.Wait()
		close(out)
	}()

	return out
}

func main() {
	ch1 := producer1()
	ch2 := producer2()

	var buffer []<-chan int
	buffer = append(buffer, ch1, ch2)

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		for v := range consumer(context.Background(), buffer...) {
			println(v)
		}
	}()

	wg.Wait()
}
