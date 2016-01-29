echo install bgpdump

mkdir -p /tmp/bgpdump-build
cd /tmp/bgpdump-build
curl http://www.ris.ripe.net/source/bgpdump/libbgpdump-1.4.99.15.tgz -o libbgpdump-1.4.99.15.tgz
tar -xf libbgpdump-1.4.99.15.tgz
cd libbgpdump-1.4.99.15
./bootstrap.sh
make
sudo make install
rm -rf /tmp/bgpdump-build
