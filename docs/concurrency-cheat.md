#### 1. WaitGroup (basic sync)
```Golang
var wg sync.WaitGroup

wg.Add(1)
go func() {
    defer wg.Done()
    // work
}()
wg.Wait() // blocks until all Done() called
```

#### 2. Collect Results with Channel
```Golang
results := make(chan string)

wg.Add(1)
go func() {
    defer wg.Done()
    results <- "done"
}()

go func() {
    wg.Wait()
    close(results)
}()

for res := range results {
    fmt.Println(res)
}
```

#### 3. Worker Pool (limit concurrency)
```Golang
jobs := make(chan int, 100)
results := make(chan int, 100)

for w := 1; w <= 5; w++ { // 5 workers
    go func(id int) {
        for j := range jobs {
            results <- j * 2 // work
        }
    }(w)
}

for j := 1; j <= 10; j++ { jobs <- j }
close(jobs)

for a := 1; a <= 10; a++ { fmt.Println(<-results) }
```

#### 4. Fail-Fast with Context
```Golang
ctx, cancel := context.WithCancel(context.Background())

go func() {
    select {
    case <-ctx.Done():
        return // stop early
    default:
        // work
        if err != nil {
            cancel() // stop everyone
        }
    }
}()
```

#### 5. Retries
```Golang
maxRetries := 3
for attempt := 1; attempt <= maxRetries; attempt++ {
    err := doWork()
    if err == nil {
        break
    }
    if attempt < maxRetries {
        time.Sleep(200 * time.Millisecond)
    }
}
```

#### 6. Exponential Backoff
```Golang
base := 100 * time.Millisecond
for attempt := 1; attempt <= 5; attempt++ {
    err := doWork()
    if err == nil { break }

    sleep := base * (1 << (attempt-1)) // 100,200,400...
    time.Sleep(sleep)
}
```

⚡ With just these 6, you can handle 90% of real-world concurrency cases in Go.
Next step (when you’re ready): jitter + rate limiting patterns.