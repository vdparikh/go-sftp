# go-sftp
The project leverages GoLang SFTP package github.com/pkg/sftp to download remote files locally. It also sets up the CRON schedule for downloading the files and renaming it.

### Build
```
go build
```

### Creating Release
Install GoReleaser prior to running below commands https://goreleaser.com
```
git tag -a v0.1.0 -m "Initial release"
git push origin v0.1.0
$GOPATH/bin/goreleaser
```
