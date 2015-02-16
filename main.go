package main

import (
	"flag"
	"fmt"
	"os"
	"runtime"
	"time"
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

	uploader, err := NewSwiftFileUploader(SwiftFileUploaderOptions{
		IdentityEndpoint: *identityEndpoint,
		Username:         *username,
		Password:         *password,
		TenantID:         *tenantId,
		SwiftRegion:      *swiftRegion,
		SwiftService:     *swiftService,
		ContainerName:    *containerName,
		Verbose:          *verbose,
	})
	if err != nil {
		fmt.Println("Cannot connect to the Swift service")
		if *verbose {
			fmt.Println(err)
		}
		os.Exit(2)
	}

	p := NewPool(runtime.NumCPU() * 5)
	watcher = NewWatcher(p)

	fmt.Println("Waiting for new files...")

	watcher.Watch(*dir, *pollInterval, func(path string) Job {
		return uploader.NewJobForFile(path)
	})
}
