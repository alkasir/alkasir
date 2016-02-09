echo install go

GO_VERSION=1.5.3
GO_GOOS=linux
GO_GOARCH=amd64
mkdir -p /tmp/go
curl -sSL https://golang.org/dl/go$GO_VERSION.$GO_GOOS-$GO_GOARCH.tar.gz | tar -C /tmp/go -xz
curl -sSL https://golang.org/dl/go${GO_VERSION}.src.tar.gz | tar -C /usr/local -xz && mkdir -p /go/bin
cd /usr/local/go/src
GOROOT_BOOTSTRAP=/tmp/go/go ./make.bash --no-clean 2>&1
rm -rf /tmp/go
