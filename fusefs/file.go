package fusefs

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"bazil.org/fuse"
)

// File represents a file node in the FUSE filesystem.
// Path is a relative path under NFSRoot (e.g., "projects/foo.txt").
type File struct {
	Path string
}

// Attr sets metadata for this file node, including read-only permissions and size.
// This method is called by the kernel to populate file details like size and mode (e.g., for `ls -l`).
func (f *File) Attr(ctx context.Context, a *fuse.Attr) error {
	nfsPath := filepath.Join(NFSRoot, f.Path)
	info, err := os.Stat(nfsPath)
	if err != nil {
		return err
	}
	// read-only regular file
	a.Mode = 0444
	a.Size = uint64(info.Size())
	return nil
}

// ReadAll returns the full contents of the file, possibly served from a content-addressed cache.
// If the file has not changed since last cache, the cached data is returned immediately.
// Otherwise, the file is read from NFSRoot (simulating slowness), cached, and then returned.
func (f *File) ReadAll(ctx context.Context) ([]byte, error) {
	log.Printf("📁 Requesting file %s", f.Path)

	// Fetch fileInfo to check last modification time
	nfsPath := filepath.Join(NFSRoot, f.Path)
	fileInfo, err := os.Stat(nfsPath)

	if err != nil {
		return nil, fmt.Errorf("%s not found in NFS", f.Path)
	}

	if fileInfo.IsDir() {
		return nil, fmt.Errorf("%s is a directory, not a file", nfsPath)
	}

	// If metadata exists, cache file exists, and timestamp matches, read from cache
	if entry, ok := metadataIndex[f.Path]; ok {
		if fileInfo.ModTime().Equal(entry.Timestamp) {
			cachePath := filepath.Join(SSDCache, entry.Hash)
			if cachedData, err := os.ReadFile(cachePath); err == nil {
				log.Printf("✅ Using valid cache for %s", f.Path)
				cacheTracker.Touch(entry.Hash)
				return cachedData, nil
			}
		}
	}

	// Simulate slower NFS read delay
	time.Sleep(500 * time.Millisecond)

	data, err := os.ReadFile(nfsPath)
	if err != nil {
		return nil, err
	}

	hash := hashContent(data)
	cachePath := filepath.Join(SSDCache, hash)

	err = os.MkdirAll(filepath.Dir(cachePath), 0755)
	if err != nil {
		return nil, err
	}

	if _, err := os.Stat(cachePath); os.IsNotExist(err) {
		err = os.WriteFile(cachePath, data, 0644)
		if err != nil {
			return nil, err
		}
		log.Printf("💾 Cached new file: %s", cachePath)
	} else {
		// When a file gets reverted, the file hash may match an old cache instance, but with a new timestamp
		log.Printf("📦 Reusing existing cache file: %s", cachePath)
	}

	metadataIndex[f.Path] = CacheEntry{
		Hash:      hash,
		Timestamp: fileInfo.ModTime(),
	}
	saveMetadataToFile()
	cacheTracker.Touch(hash)
	return data, nil
}
