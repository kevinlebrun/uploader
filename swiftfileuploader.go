package main

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"

	"github.com/rackspace/gophercloud"
	"github.com/rackspace/gophercloud/openstack"
	"github.com/rackspace/gophercloud/openstack/objectstorage/v1/objects"
)

type SwiftFileUploadJob struct {
	Path     string
	uploader *SwiftFileUploader
}

func (f *SwiftFileUploadJob) Id() string {
	return f.Path
}

func (f *SwiftFileUploadJob) Execute() {
	if f.uploader.Verbose {
		fmt.Printf("Handling %q\n", f.Path)
	}

	if !f.uploader.isExists(f.Path) {
		file, err := os.Open(f.Path)
		if err != nil {
			if f.uploader.Verbose {
				fmt.Printf("Error reading file %q\n", err)
			}
			return
		}
		err = f.uploader.create(f.Path, file)
		if err != nil {
			if f.uploader.Verbose {
				fmt.Printf("Error during upload %q\n", err)
			}
			return
		}
	}

	os.Remove(f.Path)
}

type SwiftFileUploader struct {
	Client        *gophercloud.ServiceClient
	ContainerName string
	Verbose       bool
}

func (u *SwiftFileUploader) NewJobForFile(path string) Job {
	return &SwiftFileUploadJob{
		Path:     path,
		uploader: u,
	}
}

func (u *SwiftFileUploader) KeyFromPath(filepath string) string {
	return path.Base(filepath)
}

func (u *SwiftFileUploader) create(filepath string, content io.Reader) error {
	res := objects.Create(u.Client, u.ContainerName, u.KeyFromPath(filepath), content, objects.CreateOpts{})
	if u.Verbose {
		fmt.Printf("File %q uploaded\n", u.KeyFromPath(filepath))
	}
	return res.Err
}

func (u *SwiftFileUploader) isExists(filepath string) bool {
	signature, err := u.getRemoteFileSignature(u.KeyFromPath(filepath))
	if err == nil && signature == u.getFileSignature(filepath) {
		return true
	}
	return false
}

func (u *SwiftFileUploader) getRemoteFileSignature(name string) (string, error) {
	res := objects.Get(u.Client, u.ContainerName, name, objects.GetOpts{})
	if res.Err == nil {
		return res.Header.Get("Etag"), nil
	}
	return "", res.Err
}

func (u *SwiftFileUploader) getFileSignature(filepath string) string {
	content, _ := ioutil.ReadFile(filepath)
	sum := md5.Sum(content)
	return hex.EncodeToString(sum[:])
}

type SwiftFileUploaderOptions struct {
	IdentityEndpoint string
	Username         string
	Password         string
	TenantID         string
	SwiftRegion      string
	SwiftService     string
	ContainerName    string
	Verbose          bool
}

func NewSwiftFileUploader(opts SwiftFileUploaderOptions) (*SwiftFileUploader, error) {
	provider, err := openstack.AuthenticatedClient(gophercloud.AuthOptions{
		IdentityEndpoint: opts.IdentityEndpoint,
		Username:         opts.Username,
		Password:         opts.Password,
		TenantID:         opts.TenantID,
	})
	if err != nil {
		return nil, err
	}

	client, err := openstack.NewObjectStorageV1(provider, gophercloud.EndpointOpts{
		Region: opts.SwiftRegion,
		Name:   opts.SwiftService,
	})
	if err != nil {
		return nil, err
	}

	return &SwiftFileUploader{
		Client:        client,
		ContainerName: opts.ContainerName,
		Verbose:       opts.Verbose,
	}, nil
}
