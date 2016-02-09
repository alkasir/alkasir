echo setup vagrant env

cat <<- 'EOF' > /etc/profile.d/gopath.sh
export GOPATH=${HOME}
export PATH=/usr/local/go/bin:$GOPATH/bin:$PATH
export GO15VENDOREXPERIMENT=1
EOF
