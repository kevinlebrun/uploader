# Go File Uploader

Watch for new files in a given directory and upload them to a remote location.

## Usage

Upload to a Swift container:

```
$ go build .
$ ./uploader -h
$ ./uploader  --dir ./test  --identity-endpoint https://auth.runabove.io/v2.0  --username <username>  --password <password> --tenant-id <tenant-id>  --swift-region <region>  --container-name test
```

You can use the provider `uploader.service` to deploy the uploader on a given box.
I personnaly use Ansible to do so.
Note that you will need a new user `uploader`.
You can change the user in the service file.

## Build

Generate release assets:

```
$ go build .
$ ./uploader -h
$ goxc -bc="linux,amd64 darwin,amd64" -pv=0.2.1
```

## License

The MIT license
