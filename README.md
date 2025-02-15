# wazerofs
Helper tools for custom Wazero FS

Forked from [karelbilek/wazero-fs-tools](https://github.com/karelbilek/wazero-fs-tools) for experimental purposes.

Go doc: https://pkg.go.dev/github.com/karelbilek/wazero-fs-tools

## memfs

MemFS is a in-memory filesystem. Note that minimal amount of functionality is actually done;
feel free to add a PR.

The actual underlying implementation is github.com/blang/vfs/memfs.
Here it's just a tiny wrapper around it.

github.com/blang/vfs/memfs seems to be no longer maintained, so if there is some issue, I can eventually
subtree it here; but I don't think it's necessary for now.

## sysfs

SysFS is just a verbatim copy of wazero internal sysfs. Useful for mixing with wraplogfs.

Do NOT use it in real-life code, because it is not being kept up-to-date with wazero; use it only for experimenting/debugging.

## wraplogfs

WrapLogFS is a wrapper around existing filesystem that logs all inputs/outputs

# Example - log FS

```go
import (
    "github.com/tetratelabs/wazero"
    // note: either this or wazero-fs-tools/sysfs needs to be renamed in import
    expsysfs "github.com/tetratelabs/wazero/experimental/sysfs"
    
    "github.com/gaukas/wazerofs/sysfs"
    "github.com/gaukas/wazerofs/wraplogfs"
)

// ...
func main() {
    rootFS := sysfs.DirFS("/")
    wrappedFS := wraplogfs.New(rootFS, os.Stdout, false, "root fs")

    fsConfig := wazero.NewFSConfig()
    fsConfig = fsConfig.(expsysfs.FSConfig).WithSysFSMount(wrappedFS, "/") 
    // now all / file operations will be logged

    moduleConfig := wazero.NewModuleConfig().WithFSConfig(fsConfig)./*...*/
}
```

# Example - memory FS

```go
import (
    "log"

    "github.com/tetratelabs/wazero"
    expsysfs "github.com/tetratelabs/wazero/experimental/sysfs"
    
    "github.com/gaukas/wazerofs/memfs"
)

// ...
func main() {
    memFS := memfs.New()

    // can write some files for start
    err := rootFS.WriteFile("tmp/foo.txt", []byte("this is content"))
    if err != nil {
        log.Fatal(err)
    }

    fsConfig := wazero.NewFSConfig()
    fsConfig = fsConfig.(expsysfs.FSConfig).WithSysFSMount(memFS, "/") 
    // all now happens in memory
}
```