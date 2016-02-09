echo setup vagrant profile

cat <<- 'EOF' > ${HOME}/.profile
if [ -n "$BASH_VERSION" ]; then
    # include .bashrc if it exists
    if [ -f "$HOME/.bashrc" ]; then
    . "$HOME/.bashrc"
    fi
fi

# set PATH so it includes user's private bin if it exists
if [ -d "$HOME/bin" ] ; then
    PATH="$HOME/bin:$PATH"
fi

export GOPATH=${HOME}
export PATH=/usr/local/go/bin:$GOPATH/bin:$PATH
export GO15VENDOREXPERIMENT=1

[ -z "\$PS1" ] && return
cd $GOPATH/src/github.com/alkasir/alkasir  >/dev/null|| true
EOF
