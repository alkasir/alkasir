#!/bin/bash
set -e

build_template() {
  template=jstemplate.go
  rm -f $template

  cat > $template <<EOF
package pac

const pacRawTmpl = \`
EOF
  cat pacHeaderTemplate.js >> $template
  cat pacImpl.js >> $template
  echo "\`"  >> $template
}

build_tests() {
  testdir=`mktemp -d pactestsXXXX`
  testjs=$testdir/runtests.js
  cat pacHeaderTests.js >> $testjs
  cat pacImpl.js >> $testjs
  cat pacTests.js >> $testjs
  node $testjs
  rm -rf $testdir
}

cd $(dirname $0)

echo "run tests..."
build_tests
echo "rebuild template..."
build_template
