# Nic

![GitHub release](https://img.shields.io/github/release/eddieivan01/nic.svg?label=nic)  ![GitHub issues](https://img.shields.io/github/issues/eddieivan01/nic.svg)

[English](https://github.com/EddieIvan01/nic/blob/master/README.md) | 中文

Nic是一个拥有优雅易用的API的HTTP请求库

***

### 特性

+ 封装了HTTP标准库，提供了优雅易用的API
+ 通过`nic.Session`来保持session
+ 线程（go-routine）安全

***

### 安装

输入下面的命令来安装Nic

```
$ go get -v -u github.com/eddieivan01/nic
```

***

### 快速开始

像这样发送一个HTTP请求

```go
resp, err := nic.Get("http://example.com", nil)
if err != nil {
    log.Fatal(err.Error())
}
fmt.Println(resp.Text)
```

***

### 文档

#### 发起一个基本的请求

nic可以发送以下方法的请求

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

#### 带data的post请求

```go
resp, err := nic.Post(url, nic.H{
    Data : nic.KV{
        "nic" : "nic",
    },
    Headers : nic.KV{
        "X-Forwarded-For" : "127.0.0.1",
    },
})
```

#### 带cookies的请求

```go
resp, err := nic.Get(url, nic.H{
    Cookies : nic.KV{
        "cookie1" : "nic",
    },
})
```

#### 带文件的请求

```go
resp, err := nic.Post(url, nic.H{
    Files : nic.F{
        "file" : nic.KV{
            // path of file, filename will be `nic.go`
            "filename" : `/home/nic/nic.go`,
            "token" : "0xff",
        },
    },
})
```

#### 带JSON的请求

```go
resp, err := nic.Post(url, nic.H{
    JSON : nic.KV{
        "nic" : "nic",
    }
})
```

#### 发送未经编码的原生数据

```go
resp, err := nic.Post(url, nic.H{
    Raw : "post body which is unencoded",
})
```

#### 所有的参数

```go
H struct {
    Data    KV
    Raw     string
    Headers KV
    Cookies KV
    Auth    KV
    Proxy   string

    AllowRedirect bool
    Timeout       int64
    Chunked       bool

    JSON  KV
    Files F
}
```

#### 注意!!!

`nic.H` 只能带有以下四种参数的一个

`H.Raw, H.Data, H.Files, H.JSON`

#### 用session发起请求，session可以保存服务器的`set-cookie`选项设置的cookie

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

#### 处理响应

```go
resp, _ := nic.Get(url, nil)
fmt.Println(resp.Text)
fmt.Println(resp.Bytes)
```

#### 处理JSON响应

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

#### 改变响应的编码

如果编码改变了的话，`SetEncode` 函数每一次调用都会把`resp.Bytes`转换到`resp.Text`

```go
resp, _ := nil.Get(url, nil)
err := resp.SetEncode("gbk")

if err == nil {
    fmt.Println(resp.Text)
}
```

***

### QA

+ Q:

  如何从`nic.Session`中获得原始的`*http.Request`?

  A:

  通过 `nic.Session.GetRequest` 方法

+ Q:

  如何通过 `nic.Response`将原始的 `*http.Response` 传递给类似于goquery的DOM解析库?

  A:

  通过`resp, _ := nic.Get(...); resp.Response` 来访问原始的匿名结构体字段`*http.Response`; 而且nic中`(*http.Response).Body's IO.Reader` 的bytes被拷贝了, 你可以像使用原始的 `*http.Response` 一样来使用它

+ Q:

  默认只允许十次重定向，我如何增加这个次数?

  A:

  通过访问 `nic.Session.Client` 然后修改它的CheckRedirect属性

+ Q:

  如何使用chunked传输编码

  A:

  通过设置nic.H{Chunked: true}