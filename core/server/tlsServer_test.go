package server

import (
	"bytes"
	"fmt"
	"mime"
	"net/http"
	"os"
	"testing"
)

func TestTrojanServer(t *testing.T) {

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
