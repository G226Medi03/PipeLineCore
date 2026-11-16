# PipeLineCore

Application pipeline and node model.

```go
pipeline := core.Pipeline{
    Name: "example",
    Nodes: []core.Node{
        core.NewNode(
            "load",
            "prepare",
            core.KindPrepare,
            core.CacheInherit,
            load,
            false,
        ),
        core.NewNode(
            "explain",
            "explain",
            core.KindLLM,
            core.CacheUse,
            explain,
            false,
        ),
    },
}
err := pipeline.Run(ctx, core.NewState())
```

## Cache mode

`CacheMode` is owned by this package.

- `CacheInherit`: use the run-level mode
- `CacheUse`: read and write the cache
- `CacheBypass`: skip cache reads, write a normal entry
- `CacheRefresh`: skip cache reads and refresh the existing entry

`Node.CacheModeOr(runMode)` resolves the node override.

## Construction

`core.Node` must be created with `core.NewNode`. Direct composite literals are
rejected by the `nodeconstructor` golangci-lint module plugin.

## Lint

Lint runs locally, not in GitHub Actions. Run it before every push:

```powershell
scripts/lint.ps1
```

Enable the repository pre-push hook once per clone:

```powershell
scripts/install-git-hooks.ps1
```

The hook runs the same lint script, so a failing lint blocks the push.
