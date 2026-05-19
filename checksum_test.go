package cinc

import "testing"

func TestMD5Hex(t *testing.T) {
	// md5("") is a well-known constant.
	if got := md5Hex([]byte("")); got != "d41d8cd98f00b204e9800998ecf8427e" {
		t.Fatalf("md5Hex empty = %q", got)
	}
}

func TestMD5Base64(t *testing.T) {
	// base64(md5("")) — Chef sandboxes use base64-encoded MD5.
	if got := md5Base64([]byte("")); got != "1B2M2Y8AsgTpgAmY7PhCfg==" {
		t.Fatalf("md5Base64 empty = %q", got)
	}
}
