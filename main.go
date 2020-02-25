package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"sync"
)

const (
	word = "Go"
	k    = 5
)

type result struct {
	url   string
	count int
	err   error
}

func sendMessage(ch <-chan result, sum *int, mutex *sync.Mutex) {
	for c := range ch {
		url, count, err := c.url, c.count, c.err

		if err != nil {
			fmt.Printf("Problem with `%s`: %s\n", url, err)
			continue
		}
		mutex.Lock()
		*sum += count
		fmt.Printf("Count for %s: %d\n", url, count)
		mutex.Unlock()
	}
}

func countWords(url string) (result result) {
	result.url = url

	resp, err := http.Get(url)
	if err != nil {
		result.err = err
		return
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		result.err = err
		return
	}

	result.count = strings.Count(string(body), word)

	return
}

func main() {
	var sum int
	mutex := new(sync.Mutex)
	wg := new(sync.WaitGroup)
	url := make(chan result)
	sem := make(chan struct{}, k)
	defer close(url)

	fmt.Println("Введите коректные адреса")
	fmt.Println("Введите \"end\" для подсчета суммы")

	go sendMessage(url, &sum, mutex)

	scanner := bufio.NewScanner(os.Stdin)
	scanner.Split(bufio.ScanWords)

	for scanner.Scan() {
		text := scanner.Text()
		if text == "end" {
			break
		}
		sem <- struct{}{}
		wg.Add(1)
		go func(text string, wg *sync.WaitGroup) {
			defer wg.Done()
			url <- countWords(text)
			<-sem
		}(text, wg)
	}
	wg.Wait()
	mutex.Lock()
	fmt.Println("Total:", sum)
	mutex.Unlock()
}
