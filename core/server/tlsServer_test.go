package server

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"github.com/cyejing/shuttle/pkg/codec"
	config "github.com/cyejing/shuttle/pkg/config/client"
	"io"
	"mime"
	"net/http"
	"os"
	"testing"
)

func TestTrojanServer(t *testing.T) {
	config.GlobalConfig = &config.Config{
		RemoteAddr: "s.cyejing.cn:4843",
		Password:   "123",
	}

	addr, err := codec.NewAddressFromAddr("tcp", "localhost:8088")
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	conn, err := codec.DialTrojan(&codec.Metadata{
		Command: 0,
		Address: addr,
	})
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	bs, err := io.ReadAll(conn)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	fmt.Println(hex.Dump(bs))
	fmt.Println(string(bs))

}

func TestName(t *testing.T) {
	ty := mime.TypeByExtension(".html")
	fmt.Printf("%v \n", ty)

	fb, err := os.ReadFile("/Users/chenyejing/Project/shuttle/example/index.html")
	if err != nil {
		return
	}
	contentType := http.DetectContentType(fb)
	fmt.Println(contentType)
	fmt.Printf("%s \n", fb)

	var html404 = "<html>\n<head><title>404 Not Found</title></head>\n<body>\n<center><h1>404 Not Found</h1></center>\n<hr>\n</body>\n</html>"

	buf := bytes.NewBufferString(html404)
	sc := http.DetectContentType(buf.Bytes())
	fmt.Println(sc)

}
