package encrypt

import (
	"testing"
)

type TestStruct struct {
	Msg    string
	Status bool
}

func TestInvalidEncryptBase64(t *testing.T) {
	in := TestStruct{"Hello world!", true}
	_, e := EncryptBase64("aes", "too_short", &in)
	if e.Error() == "IV must be random 32-chars" {
		return
	}

	t.Fatal(e)
}

func TestValidEncryptBase64(t *testing.T) {
	iv := "12345678912345678912345678900000"
	in := TestStruct{"Hello world!", true}
	s, e := EncryptBase64("aes", iv, &in)
	if e != nil {
		t.Fatal(e)
	}

	res := TestStruct{}
	if e := DecryptBase64("aes", iv, s, &res); e != nil {
		t.Fatal(e)
	}

	if res.Msg != "Hello world!" {
		t.Errorf("res.Msg doesn't match: " + res.Msg)
	}
	if !res.Status {
		t.Errorf("t.Status doesn't match")
	}

}