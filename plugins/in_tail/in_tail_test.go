package in_tail

import (
	"testing"
)

func TestTrimCrLf(t *testing.T) {
	if trimCrLf("hoge") != "hoge" {
		t.Fail()
	}
	if trimCrLf("hoge\n") != "hoge" {
		t.Fail()
	}
	if trimCrLf("hoge\r") != "hoge" {
		t.Fail()
	}
	if trimCrLf("hoge\r\n") != "hoge" {
		t.Fail()
	}
	if trimCrLf("hoge\n\r") != "hoge" {
		t.Fail()
	}
}
