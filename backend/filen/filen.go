// Package filen provides an interface to
// Filen cloud storage.
package filen

import (
	"bytes"
	"context"
	sdk "github.com/JupiterPi/filen-sdk-go/filen"
	"github.com/rclone/rclone/fs"
	"github.com/rclone/rclone/fs/config/configmap"
	"github.com/rclone/rclone/fs/config/configstruct"
	"github.com/rclone/rclone/fs/config/obscure"
	"github.com/rclone/rclone/fs/hash"
	"io"
	pathModule "path"
	"time"
)

func init() {
	fs.Register(&fs.RegInfo{
		Name:        "filen",
		Description: "Filen",
		NewFs:       NewFs,
		Options: []fs.Option{
			{
				Name:     "email",
				Help:     "Filen account email",
				Required: true,
			},
			{
				Name:       "password",
				Help:       "Filen account password",
				Required:   true,
				IsPassword: true,
				Sensitive:  true,
			},
		},
	})
}

func NewFs(_ context.Context, name, root string, m configmap.Mapper) (fs.Fs, error) {
	opt := new(Options)
	err := configstruct.Set(m, opt)
	if err != nil {
		return nil, err
	}

	password, err := obscure.Reveal(opt.Password)
	filen, err := sdk.New(opt.Email, password)
	if err != nil {
		return nil, err
	}
	return &Fs{name, root, filen}, nil
}

type Fs struct {
	name  string
	root  string
	filen *sdk.Filen
}

func (f *Fs) resolvePath(path string) string {
	return pathModule.Join(f.root, path)
}

type Options struct {
	Email    string `config:"email"`
	Password string `config:"password"`
}

func (f *Fs) Name() string {
	return f.name
}

func (f *Fs) Root() string {
	return f.root
}

func (f *Fs) String() string {
	return "Filen to String"
}

func (f *Fs) Precision() time.Duration {
	return 1 * time.Second
}

func (f *Fs) Hashes() hash.Set {
	return 0
}

func (f *Fs) Features() *fs.Features {
	return &fs.Features{
		//TODO implement
	}
}

func (f *Fs) List(ctx context.Context, dir string) (entries fs.DirEntries, err error) {
	dirUUID, err := f.filen.PathToUUID(f.resolvePath(dir), true)
	if err != nil {
		return nil, err
	}

	files, directories, err := f.filen.ReadDirectory(dirUUID)
	if err != nil {
		return nil, err
	}

	for _, directory := range directories {
		entries = append(entries, &Directory{
			fs:      f,
			id:      directory.UUID,
			path:    pathModule.Join(dir, directory.Name),
			size:    -1,
			items:   -1,
			created: directory.Created,
		})
	}
	for _, file := range files {
		entries = append(entries, &File{
			fs:   f,
			file: file,
			name: pathModule.Join(dir, file.Name),
		})
	}
	return entries, nil
}

type Directory struct {
	fs      *Fs
	id      string
	path    string
	size    int64
	items   int64
	created time.Time
}

func (dir *Directory) Fs() fs.Info {
	return dir.fs
}

func (dir *Directory) String() string {
	return dir.path //TODO tmp
}

func (dir *Directory) Remote() string {
	return dir.path
}

func (dir *Directory) ModTime(ctx context.Context) time.Time {
	return dir.created //TODO best guess?
}

func (dir *Directory) Size() int64 {
	return dir.size
}

func (dir *Directory) Items() int64 {
	return dir.items
}

func (dir *Directory) ID() string {
	return dir.id
}

//TODO refactor Directory and File into one DirEntry?

type File struct {
	fs   *Fs
	file *sdk.File
	name string
}

func (file *File) Fs() fs.Info {
	return file.fs
}

func (file *File) String() string {
	return file.name
}

func (file *File) Remote() string {
	return file.name
}

func (file *File) ModTime(ctx context.Context) time.Time {
	return file.file.LastModified
}

func (file *File) Size() int64 {
	return file.file.Size
}

func (file *File) Hash(ctx context.Context, ty hash.Type) (string, error) {
	return "", nil //TODO tmp
}

func (file *File) Storable() bool {
	return true
}

func (file *File) SetModTime(ctx context.Context, t time.Time) error {
	return nil //TODO tmp
}

func (file *File) Open(ctx context.Context, options ...fs.OpenOption) (io.ReadCloser, error) {
	content, err := file.fs.filen.DownloadFileInMemory(file.file)
	if err != nil {
		return nil, err
	}
	return io.NopCloser(bytes.NewBuffer(content)), nil
}

func (file *File) Update(ctx context.Context, in io.Reader, src fs.ObjectInfo, options ...fs.OpenOption) error {
	return nil //TODO tmp
}

func (file *File) Remove(ctx context.Context) error {
	return nil //TODO tmp
}

func (f *Fs) NewObject(ctx context.Context, remote string) (fs.Object, error) {
	return nil, nil
}

func (f *Fs) Put(ctx context.Context, in io.Reader, src fs.ObjectInfo, options ...fs.OpenOption) (fs.Object, error) {
	return nil, nil
}

func (f *Fs) Mkdir(ctx context.Context, dir string) error {
	return nil
}

func (f *Fs) Rmdir(ctx context.Context, dir string) error {
	return nil
}
