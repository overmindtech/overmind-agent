# Network Discovery

Currently we have an SS provider for network discover but this shhould probably be replaced with a native Go implementation as this won't work in containers since they are very unlikely to have ss installed.

If we were to build a Go implementation we would work by looking at `/proc/net/tcp` etc. for the socketys that are liening:

```
```

Then cross referencing the `inode` with the symlinks that are found in `/proc/{pid}/fd`. We will need to scan this directory anyway in order to gather process info if we were doing a pure go implementation of process gathering. A good example of this that we might be able to work from is: https://github.com/wheelcomplex/lsof