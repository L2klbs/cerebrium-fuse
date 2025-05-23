package fusefs

import (
	"log"
	"bazil.org/fuse/fs"
)

// FS implements the fs.FS interface and represents the entire mounted filesystem.
// Root is called once after mounting to provide the top-level directory node.
type FS struct{}

func (FS) Root() (fs.Node, error) {
	log.Println("File System - NFS Root called")
	return &Dir{Path: ""}, nil
}