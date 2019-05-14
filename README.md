# nic

Nic is a HTTP request library designed to send a HTTP request easier

***

### Installation

To install nic, enter the following command

```
$ go get -v -u github.com/eddieivan01/nic
```

***

### Quick start

Do a HTTP request like this

```go
resp, err := nic.Get("http://example.com", nil)
if err != nil {
    log.Fatal(err.Error())
}
fmt.Println(resp.Text)
```

***

### Documentation

**do a basic request**

nic could do these methods' request

`"HEAD", "GET", "POST", "DELETE", "OPTIONS", "PUT", "PATCH", "CONNECT", "TRACE"`

```go
import (
	"fmt"
    "github.com/eddieivan01/nic"
)

func main() {
    url := "http://example.com"
    resp, err := nic.Get(url, nil)
    if err != nil {
        log.Fatal(err.Error())
    }
    fmt.Println(resp.Text)
}
```

**post request with some data**

```go
resp, err := nic.Post(url, &nic.H{
    Data : nic.KV{
        "nic" : "nic",
    },
    Headers : nic.KV{
        "X-Forwarded-For" : "127.0.0.1",
    },
})
```

**request with cookies**

```go
resp, err := nic.Get(url, &nic.H{
    Cookies : nic.KV{
        "cookie1" : "nic",
    },
})
```

**request with files**

```go
resp, err := nic.Post(url, &nic.H{
    Files : nic.F{
        "file" : nic.KV{
            // path of file, filename will be `nic.go`
            "filename" : `/home/nic/nic.go`,
            "token" : "0xff",
        },
    },
})
```

**request with JSON**

```go
resp, err := nic.Post(url, &nic.H{
    JSON : nic.KV{
        "nic" : "nic",
    }
})
```

**request with unencoded message**

```go
resp, err := nic.Post(url, &nic.H{
    Raw : "post body which is unencoded",
})
```

**all the parameters**

```go
H struct {
    AllowRedirect bool
    Timeout       int64
    Data          KV
    Raw           string
    Headers       KV
    Cookies       KV
    Auth          KV
    Proxy         string

    JSON  KV
    Files F
}
```

**NOTICE!!!**

`nic.H` can only have one of the following four parameters

`H.Raw, H.Data, H.Files, H.JSON`

**request with session, which could save server's`set-cookie` header**

```go
session := &nic.Session{}
resp, err := session.Post("http://example.com/login", nic.H{
    Data : nic.KV{
        "uname" : "nic",
        "passwd" : "nic",
    },
})

// ......

resp, err = session.Get("http://example.com/userinfo", nil)
```

**handle response**

```go
resp, _ := nic.Get(url, nil)
fmt.Println(resp.Text)
fmt.Println(resp.Bytes)
```

**handle JSON response**

```go
resp, _ := nil.Get(url, nil)

type S struct {
    P1 string `json:"p1"`
    P2 string `json:"p2"`
}

s := &S{}
err := resp.JSON(&s)

if err == nil {
    fmt.Println(s.P1, s.P2)
}
```

**change response's encoding**

`SetEncode` will convert resp.Bytes to resp.Text if encoding is changed every time be called 

```go
resp, _ := nil.Get(url, nil)
err := resp.SetEncode("gbk")

if err == nil {
    fmt.Println(resp.Text)
}
```

