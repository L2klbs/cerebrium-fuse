# cerebrium-fuse: FUSE-based Read-through Caching Filesystem

This project implements a FUSE-based read-only filesystem that mirrors files from a slow "NFS" backend and caches them to a fast "SSD" layer. It is designed for performance-conscious environments, enabling faster repeated reads by caching file contents intelligently using content hashes and a least-recently-used (LRU) eviction policy.

## What It Does

- Mounts a virtual filesystem at `/mnt/all-projects`
- Mirrors a real directory tree from `./nfs` (slow origin)
- Caches files in `./ssd` (fast layer)
- On first read, files are fetched from NFS with a 500ms delay
- On subsequent reads, files are served instantly from SSD
- Shared cache across all projects
- Enforces a maximum of 10 cached files using LRU eviction
- Uses content hashing to ensure the cache is content-aware

---

## Setup

### Linux

```bash
# Fuse Dependencies
sudo apt update
sudo apt install fuse3 -y

# Install Go 1.24.3
# You can do a direct install with a tool like brew or use a package version manager like asdf
brew install go@1.24.3

# clone project to a project directory
mkdir ~/projects
cd ~/projects
git clone git@github.com:L2klbs/cerebrium-fuse.git
cd cerebrium-fuse

# tidy go dependencies
go mod tidy

# run go application
go run ./main.go
```

### Testing the Filesystem

You can test the caching behavior manually or via the included script. These steps will confirm that your filesystem reads from NFS on first access, then from SSD cache on subsequent accesses — and demonstrates LRU eviction when the cache exceeds 10 files. Once the application is running, open a separate terminal to test out the cache. Two sample projects (greetings and calculator) are included with basic Go programs.
You can run `ls`, `cat`, and even execute Go files directly:

```bash
# navigate to mounted path
cd /mnt/all-projects/calculator

# list existing files
ls
add.go divide.go main.go multiple.go subtract.go

# read add.go - reads from NFS and saves to cache
cat add.go

# run add.go - runs from cache
go run ./add.go 3 5

# look at saved cache
ls ~/projects/cerebrium-fuse/ssd
cat ~/projects/cerebrium-fuse/cache_metadata.json
```

### Behavior Summary
* The first time you run `cat ./add.go`, the process will be slow due to the 500ms simulated delay from the NFS source. The file will be hashed and saved to SSD cache.
* When you run `go run ./add.go`, the content will be served from SSD cache — much faster.
* The caching behavior is logged in the terminal, showing whether the file was read from NFS or SSD.
* You can inspect the `./ssd` directory to see cached files, and the `cache_metadata.json` file tracks cache entries and timestamps.
* Timestamps are used to detect file changes. If a file in NFS has been modified since it was last cached, a new content hash is generated and a fresh cache entry is created.

### LRU Demonstration
To quickly populate your cache and see LRU eviction in action:

```bash
./test/populate.sh
```

This script will read all Go files in the NFS project directories. Watch the service logs to observe cache additions and evictions as the cache limit is reached.

### Running with Docker 🐳
You can run the project inside a container:

```bash
docker build -t cerebriumfs .

# Run the application in one terminal
docker run --rm -it --cap-add SYS_ADMIN --device /dev/fuse --name cerebriumfs cerebriumfs

# Exec into running running container and run your test
docker exec -it cerebriumfs bash

# add add.go to cache
cat /mnt/all-projects/calculator/add.go

# execute add.go, will run from ssd cache
go run /mnt/all-projects/calculator/add.go 3 5

# look at saved cache
ls /app/ssd
cat /app/cache_metadata.json
```

This will build and mount the filesystem within a containerized Linux environment, suitable for testing on systems that do not support FUSE natively.

### Filesystem Design

The filesystem is implemented using the bazil.org/fuse library in Go. The mount point is a virtual read-only view backed by two real directories:

* nfs/ simulates slow storage with a 500ms artificial delay per read
* ssd/ acts as a content-addressed cache for faster subsequent access

Key components:
* main.go: mounts the filesystem and runs the server loop
* fusefs/dir.go: handles directory lookups and listing
* fusefs/file.go: handles file reads and caching logic
* fusefs/cache.go: manages an in-memory LRU cache to enforce a 10-file limit
* fusefs/metadata.go: persists file hashes and timestamps to cache_metadata.json
* fusefs/config.go: defines NFSRoot and SSDCache paths

### Real-World Use Cases
* Remote development environments: Speed up file access over SSH by caching source files locally.
* Cloud IDEs (e.g., VS Code SSH): Reduce latency by caching frequently accessed project files from remote hosts.
* CDN or edge node preloading: Cache config or static assets from central storage to edge devices for low-latency delivery.
* Read-through S3 filesystems: Mirror S3 buckets locally using FUSE, cache reads to avoid costly repeated fetches.
* CI/CD pipeline caching: Cache common files across build steps to speed up test runs.

### Future Improvements
Some areas for future work and improvement include:
* Supporting a byte-based cache size limit (e.g., 100 KB)
* Adding time-to-live (TTL) expiration for rarely accessed files
* Tracking usage metrics or exposing observability endpoints
    * Utilize metrics to adjust eviction policies based on frequency
* Smart pre-fetching and caching of files (e.g. cache page 4 of a book when a user is reading page 3)
* Support multi-user environments by tracking and preserving file UID/GID/mode, allowing shared cache access with proper permissions