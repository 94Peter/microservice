package grpc_tool_test

import (
	"fmt"
	"net/url"
	"regexp"
	"testing"
)

func TestGetFromCtx(t *testing.T) {
	re := regexp.MustCompile(`^grpc(s)?://`)
	if re.MatchString("grpc://127.0.0.1:8080") {
		fmt.Println("okok")
	}

	if !re.MatchString("127.0.0.1:8080") {
		fmt.Println("OKOKokok")
	}
	u1, err := url.Parse("grpc://127.0.0.1:8080")
	fmt.Println(u1.Scheme, u1.Host, err)
	u2, err := url.Parse("grpcs://127.0.0.1")
	fmt.Println(u2.Host, u2.Scheme, u2.Opaque, u2.Port())
	fmt.Println(u2, err)
	fmt.Println("test")
	t.Errorf("Expected result.Name to be 'test', got %s", "ddd")
}
