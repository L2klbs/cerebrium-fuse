package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	"cerebrium-fuse/fusefs"
)

func main() {
	// Set the FUSE mount point
	mountPoint := "/mnt/all-projects"
	_ = os.MkdirAll(mountPoint, 0755)

	// umount mountPoint if process is ended
	setupSignalHandler(mountPoint)

	// Mount a new FUSE filesystem at the mount point.
	// This returns a connection used to exchange messages with the kernel.
	log.Printf("Mounting FUSE filesystem at %s", mountPoint)
	conn, err := fuse.Mount(
		mountPoint,
		fuse.FSName("cerebriumfs"),   // name shown in system utilities like `mount`
		fuse.Subtype("cachefs"),      // optional descriptor for the FS type
		fuse.ReadOnly(),              // mark the filesystem as read-only
		fuse.AllowOther(),            // allow access from users other than the mounter
	)

	if err != nil {
		log.Fatalf("❌ Failed to mount FUSE connection: %v", err)
	}

	// Always close the FUSE connection when done
	defer conn.Close()

	// Serve handles kernel requests using the implementation in fusefs.FS
	// This call runs in the foreground and only returns when the filesystem is unmounted or an error occurs.
	err = fs.Serve(conn, fusefs.FS{})
	if err != nil {
		log.Fatalf("❌ FUSE serve error: %v", err)
	}
}

// If process is killed, handle umount
// e.g. ctrl+c or kill
func setupSignalHandler(mountPoint string) {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigs
		fuse.Unmount(mountPoint)
		os.Exit(0)
	}()
}
