package leetcode

import (
    "github.com/sirupsen/logrus"
    "sync"
    "testing"
)

// 协程池示例代码 https://blog.51cto.com/zhangxueliang/8368196

func TestGoroutinePool(test *testing.T) {
    type Job struct {
        id int
    }
    type Result struct {
        id   int
        done bool
    }

    numOfJobs := 10
    numOfWorkers := 3

    jobs := make(chan Job, numOfJobs)
    results := make(chan Result, numOfJobs)

    worker := func(jobs chan Job, results chan Result) {
        for job := range jobs {
            // do job
            r := Result{
                id:   job.id,
                done: true,
            }
            results <- r
        }
    }

    // 创建协程池
    wg := &sync.WaitGroup{}
    for i := 0; i < numOfWorkers; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            worker(jobs, results)
        }()
    }

    // 同步提交 job
    for i := 0; i < numOfJobs; i++ {
        jobs <- Job{
            id: i + 1,
        }
    }
    close(jobs)

    go func() {
        wg.Wait()
        close(results)
    }()

    for result := range results {
        logrus.Infof("result: %+v", result)
    }
}
