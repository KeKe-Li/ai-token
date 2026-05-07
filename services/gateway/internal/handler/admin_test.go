package handler

import "testing"

func TestParseOptionalPositiveInt64AcceptsPositiveID(t *testing.T) {
	value, ok := parseOptionalPositiveInt64("42")
	if !ok || value == nil || *value != 42 {
		t.Fatalf("正整数 ID 应被解析为过滤条件，value=%v ok=%v", value, ok)
	}
}

func TestParseOptionalPositiveInt64AllowsEmptyValue(t *testing.T) {
	value, ok := parseOptionalPositiveInt64("")
	if !ok || value != nil {
		t.Fatalf("空值应表示不过滤，value=%v ok=%v", value, ok)
	}
}

func TestParseOptionalPositiveInt64RejectsInvalidValue(t *testing.T) {
	if value, ok := parseOptionalPositiveInt64("0"); ok || value != nil {
		t.Fatalf("非正整数 ID 不应被接受，value=%v ok=%v", value, ok)
	}
	if value, ok := parseOptionalPositiveInt64("abc"); ok || value != nil {
		t.Fatalf("非数字 ID 不应被接受，value=%v ok=%v", value, ok)
	}
}
