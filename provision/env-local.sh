echo setup local env
cat <<- 'EOF' > /etc/profile.d/go.sh
export GOPATH=${HOME}
export PATH=/usr/local/go/bin:$GOPATH/bin:$PATH
export GO15VENDOREXPERIMENT=1
EOF
