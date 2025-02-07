# Wopo — the simplest generalized worker pool for Go


```sh
go get github.com/dmitysh/wopo
```

* Simplicity
* Implementation based on type parameters
* No dependencies except for testify
* **Easy to use:**
```Go 
p := wopo.NewPool(
    h, // func squareizer (context.Context, int) (int, error)
    wopo.WithWorkerCount[int, int](10),
    wopo.WithResultBufferSize[int, int](3),
    wopo.WithResultBufferSize[int, int](3),
)

p.Start()

go func() {
    for i := 0; i < 300; i++ {
        p.PushTask(ctx, i)
    }
    p.Stop()
}()

for res := range p.Results() {
    fmt.Println(res.Data, res.Err)
}
```