package coreapi

import (
	"context"
	"errors"

	ft "gx/ipfs/QmZKvJkPsgy18PStoH1aT4H98wDz8ioYoa1gpbK3W9HsY7/go-unixfs"
	uio "gx/ipfs/QmZKvJkPsgy18PStoH1aT4H98wDz8ioYoa1gpbK3W9HsY7/go-unixfs/io"
	ipld "gx/ipfs/QmcKKBwfz6FyQdHR2jsXrrF6XeSBXYL86anmWNewpFpoF5/go-ipld-format"
	dag "gx/ipfs/QmdURv6Sbob8TVW2tFFve9vcEWrSUgwPqeqnXyvYhLrkyd/go-merkledag"
	files "gx/ipfs/QmeaQRmnRog7NxLEWHP9zSTkics4cbgwBVa7q49LmBowDr/go-ipfs-files"
)

// Number to file to prefetch in directories
// TODO: should we allow setting this via context hint?
const prefetchFiles = 4

// TODO: this probably belongs in go-unixfs (and could probably replace a chunk of it's interface in the long run)

type ufsDirectory struct {
	ctx   context.Context
	dserv ipld.DAGService
	dir   uio.Directory
}

type ufsIterator struct {
	ctx   context.Context
	files chan *ipld.Link
	dserv ipld.DAGService

	curName string
	curFile files.Node

	err error
}

func (it *ufsIterator) Name() string {
	return it.curName
}

func (it *ufsIterator) Node() files.Node {
	return it.curFile
}

func (it *ufsIterator) File() files.File {
	f, _ := it.curFile.(files.File)
	return f
}

func (it *ufsIterator) Dir() files.Directory {
	d, _ := it.curFile.(files.Directory)
	return d
}

func (it *ufsIterator) Next() bool {
	l, ok := <-it.files
	if !ok {
		return false
	}

	it.curFile = nil

	nd, err := l.GetNode(it.ctx, it.dserv)
	if err != nil {
		it.err = err
		return false
	}

	it.curName = l.Name
	it.curFile, it.err = newUnixfsFile(it.ctx, it.dserv, nd)
	return it.err == nil
}

func (it *ufsIterator) Err() error {
	return it.err
}

func (d *ufsDirectory) Close() error {
	return nil
}

func (d *ufsDirectory) Entries() (files.DirIterator, error) {
	fileCh := make(chan *ipld.Link, prefetchFiles)
	go func() {
		d.dir.ForEachLink(d.ctx, func(link *ipld.Link) error {
			select {
			case fileCh <- link:
			case <-d.ctx.Done():
				return d.ctx.Err()
			}
			return nil
		})

		close(fileCh)
	}()

	return &ufsIterator{
		ctx:   d.ctx,
		files: fileCh,
		dserv: d.dserv,
	}, nil
}

func (d *ufsDirectory) Size() (int64, error) {
	return 0, files.ErrNotSupported
}

type ufsFile struct {
	uio.DagReader
}

func (f *ufsFile) Size() (int64, error) {
	return int64(f.DagReader.Size()), nil
}

func newUnixfsDir(ctx context.Context, dserv ipld.DAGService, nd ipld.Node) (files.Directory, error) {
	dir, err := uio.NewDirectoryFromNode(dserv, nd)
	if err != nil {
		return nil, err
	}

	return &ufsDirectory{
		ctx:   ctx,
		dserv: dserv,

		dir: dir,
	}, nil
}

func newUnixfsFile(ctx context.Context, dserv ipld.DAGService, nd ipld.Node) (files.Node, error) {
	switch dn := nd.(type) {
	case *dag.ProtoNode:
		fsn, err := ft.FSNodeFromBytes(dn.Data())
		if err != nil {
			return nil, err
		}
		if fsn.IsDir() {
			return newUnixfsDir(ctx, dserv, nd)
		}

	case *dag.RawNode:
	default:
		return nil, errors.New("unknown node type")
	}

	dr, err := uio.NewDagReader(ctx, nd, dserv)
	if err != nil {
		return nil, err
	}

	return &ufsFile{
		DagReader: dr,
	}, nil
}

var _ files.Directory = &ufsDirectory{}
var _ files.File = &ufsFile{}
