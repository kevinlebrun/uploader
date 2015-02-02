package swiftfile

import (
	"crypto/md5"
	"encoding/hex"
	"io"
	"io/ioutil"
	"os"
	"path"

	"github.com/rackspace/gophercloud"
	"github.com/rackspace/gophercloud/openstack/objectstorage/v1/objects"
)

// TODO create a swift wrapper where the configuration lives instead of passing client to each file

type File struct {
	Path          string
	Client        *gophercloud.ServiceClient
	ContainerName string
}

func (f *File) Key() string {
	return f.Path
}

func (f *File) Upload() {
	name := path.Base(f.Path)

	if f.shouldUpdate(name) {
		file, _ := os.Open(f.Path)
		f.create(name, file)
	}

	os.Remove(f.Path)
}

func (f *File) create(name string, content io.Reader) {
	objects.Create(f.Client, f.ContainerName, name, content, objects.CreateOpts{})
}

func (f *File) shouldUpdate(name string) bool {
	signature, err := f.getRemoteFileSignature(name)
	if err == nil && signature == f.getFileSignature(f.Path) {
		return false
	}
	return true
}

func (f *File) getRemoteFileSignature(name string) (string, error) {
	res := objects.Get(f.Client, f.ContainerName, path.Base(f.Path), objects.GetOpts{})
	if res.Err == nil {
		return res.Header.Get("Etag"), nil
	}
	return "", res.Err
}

func (f *File) getFileSignature(path string) string {
	content, _ := ioutil.ReadFile(f.Path)
	sum := md5.Sum(content)
	return hex.EncodeToString(sum[:])
}
