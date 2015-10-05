package main

import (
	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	"flag"
	"fmt"
	"golang.org/x/net/context"
	"log"
	"os"
	"path/filepath"
)

var progName = filepath.Base(os.Args[0])

func usage() {
	fmt.Fprintf(os.Stderr, "Usage of %s:\n", progName)
	fmt.Fprintf(os.Stderr, "  %s DEVICE MOUNTPOINT\n", progName)
	flag.PrintDefaults()
}

func main() {
	log.SetFlags(0)
	log.SetPrefix(progName + ": ")

	flag.Usage = usage
	flag.Parse()

	if flag.NArg() != 2 {
		usage()
		os.Exit(2)
	}
	device := flag.Arg(0)
	mountpoint := flag.Arg(1)
	if err := mount(device, mountpoint); err != nil {
		log.Fatal(err)
	}
}

func mount(device, mountpoint string) error {
	c, err := fuse.Mount(mountpoint)
	scanner, _ := NewScanner(device)
	scanner.EnumerateHosts()
	if err != nil {
		return err
	}
	defer c.Close()

	filesys := &FS{scanner: scanner}
	if err := fs.Serve(c, filesys); err != nil {
		return err
	}

	<-c.Ready
	if err := c.MountError; err != nil {
		return err
	}

	return nil
}

type FS struct {
	scanner Scanner
}

var _ fs.FS = (*FS)(nil)

func (f *FS) Root() (fs.Node, error) {
	n := &Dir{scanner: f.scanner}
	return n, nil
}

type Dir struct {
	scanner Scanner
	host    Host
}

var _ fs.Node = (*Dir)(nil)

func (d *Dir) Attr(ctx context.Context, a *fuse.Attr) error {
	if d.host.address == "" {
		// root directory
		a.Mode = os.ModeDir | 0755
		return nil
	}
	a.Mode = os.ModeNamedPipe | 0755
	return nil
}

var _ = fs.HandleReadDirAller(&Dir{})

func (d *Dir) ReadDirAll(ctx context.Context) ([]fuse.Dirent, error) {
	var res []fuse.Dirent
	res = append(res, fuse.Dirent{Type: fuse.DT_Dir, Name: "."})
	res = append(res, fuse.Dirent{Type: fuse.DT_Dir, Name: ".."})
	for addr, _ := range d.scanner.Hosts {
		res = append(res, fuse.Dirent{Type: fuse.DT_File, Name: addr})
	}
	return res, nil
}

type File struct {
	host Host
}

var _ fs.Node = (*File)(nil)

func (d File) Attr(ctx context.Context, a *fuse.Attr) error {
	a.Mode = 0755
	a.Size = 1024
	return nil
}

var _ = fs.NodeRequestLookuper(&Dir{})

func (d *Dir) Lookup(ctx context.Context, req *fuse.LookupRequest, resp *fuse.LookupResponse) (fs.Node, error) {
	host, prs := d.scanner.Hosts[req.Name]
	if prs {
		return &File{host: host}, nil
	}
	return nil, fuse.ENOENT
}

var _ = fs.NodeOpener(&File{})

func (f *File) Open(ctx context.Context, req *fuse.OpenRequest, resp *fuse.OpenResponse) (fs.Handle, error) {
	f.host.nmapScan()
	return &FileHandle{host: f.host}, nil
}

type FileHandle struct {
	host Host
}

var _ fs.Handle = (*FileHandle)(nil)

var _ fs.HandleReleaser = (*FileHandle)(nil)

func (fh *FileHandle) Release(ctx context.Context, req *fuse.ReleaseRequest) error {
	return nil
}

var _ = fs.HandleReader(&FileHandle{})

func (fh *FileHandle) Read(ctx context.Context, req *fuse.ReadRequest, resp *fuse.ReadResponse) error {
	padding := make([]byte, req.Size-len(fh.host.nmapData))
	resp.Data = append(fh.host.nmapData, padding...)
	return nil
}
