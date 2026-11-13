# PipeLineCore

Application pipeline and node model.

```go
pipeline := core.Pipeline{
    Name: "example",
    Nodes: []core.Node{...},
}
err := pipeline.Run(ctx, core.NewState())
```
