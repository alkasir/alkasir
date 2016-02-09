A Go library to pull data from a [Sonatype Nexus](http://www.sonatype.com/nexus) instance.

[![TravisCI](https://travis-ci.org/hanjos/nexus.svg)](https://travis-ci.org/hanjos/nexus)
[![GoDoc](https://godoc.org/github.com/hanjos/nexus?status.svg)](https://godoc.org/github.com/hanjos/nexus)

How?
----

`go get` should see you through:

```sh
go get github.com/hanjos/nexus
```

And therefore `import`:

```Go
package main

import (
  "fmt"
  "github.com/hanjos/nexus"
  "github.com/hanjos/nexus/credentials"
  "github.com/hanjos/nexus/search"
  "reflect"
)

func main() {
  n := nexus.New("http://nexus.somewhere.com", credentials.BasicAuth("username", "password"))

  artifacts, err := n.Artifacts(
    search.InRepository{
      "shamalamadingdong",
      search.ByKeyword("com.sbrubbles*")})

  if err != nil {
    fmt.Printf("%v: %v", reflect.TypeOf(err), err)
  }

  for _, a := range artifacts {
    fmt.Println(a)
  }
}
```

Why?
----

Nexus has a large REST API, but some information isn't readily available, requiring several API calls and some mashing
up to produce.

And it was a good excuse to try Go out :)

License
-------

MIT License. See [LICENSE](https://github.com/hanjos/nexus/blob/master/LICENSE) for the gory details.
