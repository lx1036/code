/**
https://studygolang.gitbook.io/learn-go-with-tests/go-ji-chu/concurrency
 */
package concurrency

type WebsiteChecker func(string) bool
type result struct {
    string
    bool
}

func CheckWebsites(wc WebsiteChecker, urls []string) map[string]bool {
    results := make(map[string]bool)
    resultChannel := make(chan result)

    for _, url := range urls {
        //results[url] = wc(url) // go test --bench=. (2.371s)
        go func(u string) {
            //results[url] = wc(u) // go test --bench=. (0.005s)
            resultChannel <- result{u, wc(u)}
        }(url)
    }

    for i := 0; i < len(urls); i++ {
        result := <-resultChannel
        results[result.string] = result.bool // go test --bench=. BenchmarkCheckWebsites-4 55 23947639 ns/op
    }

    //time.Sleep(2 * time.Second)

    return results
}



