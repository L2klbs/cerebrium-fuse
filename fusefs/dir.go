package fusefs

import (
	"context"
	"log"
	"os"
	"path/filepath"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
)

// Dir represents a directory node in the FUSE filesystem.
// The Path field is a relative path from the configured NFS root directory.
type Dir struct {
	Path string
}

// Attr sets metadata for this directory node.
// It marks the node as a directory (os.ModeDir) and sets read+execute permissions (0555).
func (d *Dir) Attr(ctx context.Context, a *fuse.Attr) error {
	a.Mode = os.ModeDir | 0555
	return nil
}

// ReadDirAll returns all directory entries under this node.
// It lists the actual contents of the corresponding directory under NFSRoot.
// Called when a user runs `ls`, or opens the directory in a file explorer.
func (d *Dir) ReadDirAll(ctx context.Context) ([]fuse.Dirent, error) {
	log.Println("Reading directory:", d.Path)

	absPath := filepath.Join(NFSRoot, d.Path)
	files, err := os.ReadDir(absPath)

	if err != nil {
		log.Printf("❌ Failed to read directory %s", absPath)
		return nil, err
	}

	entries := []fuse.Dirent{}
	for _, file := range files {
		typeFlag := fuse.DT_File
		if file.IsDir() {
			typeFlag = fuse.DT_Dir
		}
		entries = append(entries, fuse.Dirent{
			Name: file.Name(),
			Type: typeFlag,
		})
	}

	return entries, nil
}

// Lookup resolves a child name in this directory to a file or subdirectory.
// It maps the name to a node by checking the underlying NFSRoot path on disk.
// Called when a user tries to access a file by name, like `cat file.txt`.
func (d *Dir) Lookup(ctx context.Context, name string) (fs.Node, error) {
	relPath := filepath.Join(d.Path, name)
	fullPath := filepath.Join(NFSRoot, relPath)

	log.Printf("Looking for %s under %s", name, d.Path)

	info, err := os.Stat(fullPath)
	if err != nil {
		return nil, fuse.ENOENT
	}

	if info.IsDir() {
		return &Dir{Path: relPath}, nil
	}

	return &File{Path: relPath}, nil
}
