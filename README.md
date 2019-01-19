# Filesizes

`filesizes` looks recursively in a directory for all files
over a certain size, and then prints out the files in descending
order.

## Example

```
$ filesizes -dir ~/ -min-size=400MB
Looking in "/Users/jonathanpentecost/"...
6.2 GB -> /Users/jonathanpentecost/Library/Containers/com.docker.docker/Data/vms/0/Docker.qcow2
682 MB -> /Users/jonathanpentecost/projects/src/installs/zig-macos-x86_64-0.3.0/Zig Live Coding - Parsing Floating Points, Part 2-YRxIjY5ELSM.mkv
682 MB -> /Users/jonathanpentecost/videos/Zig Live Coding - Parsing Floating Points, Part 2-YRxIjY5ELSM.mkv
587 MB -> /Users/jonathanpentecost/go/src/k8s.io/kubernetes/.git/objects/pack/pack-e3f8c78993f3392e18dd5032af40242bf5b02de9.pack
571 MB -> /Users/jonathanpentecost/go/pkg/dep/sources/https---github.com-kubernetes-kubernetes/.git/objects/pack/pack-97a964f8268f28d9bc4ec1d961214221173cff56.pack
524 MB -> /Users/jonathanpentecost/videos/Zig Live Coding - Compile Time Code Execution-mdzkTOsSxW8.webm
524 MB -> /Users/jonathanpentecost/projects/src/installs/zig-macos-x86_64-0.3.0/Zig Live Coding - Compile Time Code Execution-mdzkTOsSxW8.webm
411 MB -> /Users/jonathanpentecost/go/src/github.com/kubernetes/kubernetes/.git/objects/pack/pack-76d8961eff7c6f0b24528089b6018bf28a37650b.pack
```

```
$ filesizes -dir ~/ -min-size=1GB
Looking in "/Users/jonathanpentecost/"...
6.2 GB -> /Users/jonathanpentecost/Library/Containers/com.docker.docker/Data/vms/0/Docker.qcow2
```

## Installing

    go get -u github.com/vishen/filesizes

## Usage

```
Find files recursively from a starting directory iff the file is over a certain size.
  -dir string
    	file or directory to get file sizes from (required)
  -min-size string
    	minimum size of a file to include in results, ie: 100MB, 2GB, etc (required)
  -workers int
    	number of concurrent workers, default to runtime.GOMAXPROCS * 2 (default 8)
```

### Size

https://github.com/dustin/go-humanize is being used to parse 100MB, 1GB.
Please refer to their documentation for size conversions.

### Workers

Since the majority of the work done is by the kernel waiting for
filesystem operations, the more workers won't always be better as
there will be a lot of contention for accessing the filesystem.

The best number of workers I have found is roughly 2 * num of CPUs,
which I have set as the default.
