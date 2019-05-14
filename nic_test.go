package nic

import (
	"log"
	"strings"
	"testing"
	"time"

	"github.com/armon/go-socks5"
	"github.com/gin-gonic/gin"
)

// Testing http server addr
var baseURL = "http://127.0.0.1:2333"

func init() {
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()

	router.GET("/get", func(c *gin.Context) {
		c.String(200, "ok"+c.ClientIP()+c.GetHeader("Cookie"))
	})
	router.GET("/redirect", func(c *gin.Context) {
		c.Redirect(302, "/redirect-dst")
	})
	router.GET("/redirect-dst", func(c *gin.Context) {
		c.String(200, "redirect_ok")
	})
	router.GET("/timeout", func(c *gin.Context) {
		time.Sleep(time.Duration(3) * time.Second)
		c.String(200, "timeout")
	})
	router.GET("/cookie", func(c *gin.Context) {
		c.SetCookie("nic", "nic", 0, "/session", "127.0.0.1", true, true)
	})
	router.GET("/session", func(c *gin.Context) {
		cookie, _ := c.Cookie("nic")
		if cookie == "nic" {
			c.String(200, "session_keep_ok")
		}
	})
	authorized := router.Group("/auth", gin.BasicAuth(gin.Accounts{
		"nic": "nic"}))
	authorized.GET("/", func(c *gin.Context) {
		c.String(200, "auth ok")
	})
	router.GET("/json-resp", func(c *gin.Context) {
		type jsonStruct struct {
			P1 string `json:"p1"`
			P2 string `json:"p2"`
		}
		j := &jsonStruct{"1", "2"}
		c.JSON(200, j)
	})
	router.GET("/encode", func(c *gin.Context) {
		c.String(200, "你好")
	})

	router.POST("/data", func(c *gin.Context) {
		args1 := c.PostForm("args1")
		args2 := c.PostForm("args2")
		if args1 == "1" && args2 == "a&%%$$" {
			c.String(200, "post data ok")
		} else {
			c.String(200, "post data error")
		}
	})
	router.POST("/file", func(c *gin.Context) {
		file, _ := c.FormFile("file1")
		if file.Filename == "nic.go" {
			c.String(200, "file upload ok")
		} else {
			c.String(200, "file upload error")
		}
	})
	router.POST("/json", func(c *gin.Context) {
		type jsonStruct struct {
			P1 string `json:"p1"`
			P2 string `json:"p2"`
		}
		j := &jsonStruct{}
		c.BindJSON(j)
		if j.P1 == "11" && j.P2 == "s)(*&\"^%" {
			c.String(200, "json ok")
		} else {
			c.String(200, "json error")
		}
	})

	// run a socks5 server
	// for proxy option testing
	conf := &socks5.Config{}
	server, err := socks5.New(conf)
	if err != nil {
		log.Println(err.Error())
	}
	go server.ListenAndServe("tcp", "127.0.0.1:8088")

	go router.Run("127.0.0.1:2333")
}

// tesing via burpsuite proxy
func TestGetMethodWithNoParams(t *testing.T) {
	session := &Session{}

	resp, err := session.Request("get", baseURL+"/get", nil)
	if err != nil || resp.StatusCode != 200 {
		t.Error("get method with no params error")
	} else {
		t.Log("get method with no params ok ✔")
	}
}

func TestGetMethodWithParams(t *testing.T) {
	session := &Session{}

	resp, err := session.Request("get", baseURL+"/get", &H{
		Headers: KV{
			"X-Forwarded-For": "1.1.1.1",
		},
		Cookies: KV{
			"nic": "nic",
		},
	})
	if err != nil || !strings.Contains(resp.Text, "1.1.1.1") || !strings.Contains(resp.Text, "nic") {
		t.Error("get method with params error")
	} else {
		t.Log("get method with params ok ✔")
	}
}

func TestRedirect(t *testing.T) {
	session := &Session{}

	// modify status code by proxy
	resp1, err := session.Request("get", baseURL+"/redirect", &H{
		AllowRedirect: true,
	})
	resp2, err := session.Request("get", baseURL+"/redirect", &H{
		AllowRedirect: false,
	})
	if err != nil || resp1.StatusCode != 200 || resp1.Text != "redirect_ok" || resp2.StatusCode != 302 {
		t.Error("redirect error")
	} else {
		t.Log("redirect ok ✔")
	}
}

func TestTimeout(t *testing.T) {
	session := &Session{}

	_, err := session.Request("get", baseURL+"/timeout", &H{
		Timeout: 1,
	})
	if err != nil {
		t.Log("timeout ok ✔")
	} else {
		t.Error("timeout error")
	}
}
func TestSessionKeeping(t *testing.T) {
	session := &Session{}

	resp, _ := session.Request("get", baseURL+"/cookie", nil)
	cookies := session.cookies
	if len(cookies) == 0 || len(resp.Cookies()) == 0 {
		t.Error("session keep error")
	}
	respT, _ := session.Request("get", baseURL+"/session", nil)
	if respT.Text != "session_keep_ok" {
		t.Error("session keep error")
	}
	t.Log("session keep ok ✔")
}

func TestPostMethodWithData(t *testing.T) {
	session := &Session{}

	resp, err := session.Request("post", baseURL+"/data", &H{
		Data: KV{
			"args1": "1",
			"args2": "a&%%$$",
		},
	})
	if err != nil || resp.Text == "post data error" {
		t.Error("post method with data error ")
	} else {
		t.Log("post method with data ok ✔")
	}
}

func TestPostMethodWithFiles(t *testing.T) {
	session := &Session{}

	resp, err := session.Request("post", baseURL+"/file", &H{
		Files: F{
			"file1": KV{
				"filename": `.\nic.go`,
				"token":    "123",
			},
		},
	})
	if err != nil || resp.Text == "file upload error" {
		t.Error("post method with files error: " + err.Error())
	} else {
		t.Log("post method with files ok ✔")
	}
}

func TestPostMethodWithJson(t *testing.T) {
	session := &Session{}

	resp, err := session.Request("post", baseURL+"/json", &H{
		JSON: KV{
			"p1": "11",
			"p2": "s)(*&\"^%",
		},
	})
	if err != nil || resp.Text == "json error" {
		t.Error("post method with json error")
	} else {
		t.Log("post method with json ok ✔")
	}
}

func TestAuth(t *testing.T) {
	session := &Session{}

	resp, err := session.Request("get", baseURL+"/auth/", &H{
		Auth: KV{
			"nic": "nic",
		},
	})
	if err != nil || resp.StatusCode == 401 || resp.Text != "auth ok" {
		t.Error("auth error")
	} else {
		t.Log("auth ok ✔")
	}
}

func TestProxy(t *testing.T) {
	session := &Session{}

	resp, err := session.Request("get", baseURL+"/get", &H{
		Proxy: "socks5://127.0.0.1:8088",
	})

	if err != nil || resp.StatusCode != 200 {
		t.Error("proxy error")
	} else {
		t.Log("proxy ok ✔")
	}
}

func TestJsonParse(t *testing.T) {
	session := &Session{}

	resp, err := session.Request("get", baseURL+"/json-resp", nil)

	type jsonStruct struct {
		P1 string `json:"p1"`
		P2 string `json:"p2"`
	}
	jsonS := &jsonStruct{}
	err = resp.JSON(&jsonS)
	if err != nil || jsonS.P1 != "1" || jsonS.P2 != "2" {
		t.Error("json parse error")
	} else {
		t.Log("json parse ok ✔")
	}
}

func TestEncode(t *testing.T) {
	session := &Session{}

	resp, err := session.Request("get", baseURL+"/encode", nil)
	err = resp.SetEncode("gbk")
	if err != nil || resp.GetEncode() != "gbk" {
		t.Error("encode error")
	} else {
		t.Log("encode ok ✔")
	}
}
