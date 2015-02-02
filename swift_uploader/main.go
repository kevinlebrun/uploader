package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/kevinlebrun/uploader"
	"github.com/kevinlebrun/uploader/swiftfile"
	"github.com/rackspace/gophercloud"
	"github.com/rackspace/gophercloud/openstack"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	dir := flag.String("dir", "", "Directory where the files are located")
	containerName := flag.String("container-name", "", "SwiftContainer used to receive files")
	identityEndpoint := flag.String("identity-endpoint", "", "OpenStack identity endpoint")
	username := flag.String("username", "", "OpenStack username")
	password := flag.String("password", "", "OpenStack password")
	tenantId := flag.String("tenant-id", "", "OpenStack Tenant ID")
	swiftRegion := flag.String("swift-region", "", "OpenStack Swift region")
	swiftService := flag.String("swift-service", "swift", "OpenStack Swift service")
	pollInterval := flag.Duration("poll", 5*time.Second, "Poll interval")
	verbose := flag.Bool("verbose", false, "Show more")
	flag.Parse()

	if _, err := os.Stat(*dir); os.IsNotExist(err) {
		fmt.Printf("No such file or directory: %s\n", *dir)
		os.Exit(1)
	}

	provider, err := openstack.AuthenticatedClient(gophercloud.AuthOptions{
		IdentityEndpoint: *identityEndpoint,
		Username:         *username,
		Password:         *password,
		TenantID:         *tenantId,
	})
	if err != nil {
		fmt.Println("OpenStack authentication failed")
		if *verbose {
			fmt.Println(err)
		}
		os.Exit(2)
	}

	client, err := openstack.NewObjectStorageV1(provider, gophercloud.EndpointOpts{
		Region: *swiftRegion,
		Name:   *swiftService,
	})
	if err != nil {
		fmt.Println("Cannot connect to the Swift service")
		if *verbose {
			fmt.Println(err)
		}
		os.Exit(3)
	}

	s := uploader.NewUploader(runtime.NumCPU())

	fmt.Println("Waiting for new files...")

	go watchFiles(*dir, *pollInterval, func(path string) {
		file := &swiftfile.File{Path: path, Client: client, ContainerName: *containerName}
		if ok := s.Upload(file); ok && *verbose {
			fmt.Printf("Upload new file: %q\n", path)
		}
	})

	s.Wait()
}

func watchFiles(dir string, poll time.Duration, f func(string)) {
	for {
		filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if !info.IsDir() {
				f(path)
			}
			return nil
		})

		time.Sleep(poll)
	}
}
