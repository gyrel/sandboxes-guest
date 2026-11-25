# sandboxes-guest

<img src="docs/assets/gyrel-shark.png" alt="Gyrel shark logo" width="180">


## Import path migration

The canonical Go module path is `github.com/gyrel/sandboxes-guest`. Update imports and `go.mod` requirements when moving to a release that includes this change. Previous module paths are not kept as compatibility aliases.

Status: pre-alpha.

In-VM guest init and agent. Handles process execution, PTY, file operations, and
guest-side bootstrap. No host keychain or cloud dependencies.
