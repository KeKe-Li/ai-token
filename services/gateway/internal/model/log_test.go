package model

import (
	"strings"
	"testing"
)

func TestBuildRequestLogListQueryFiltersByRequestLogID(t *testing.T) {
	requestLogID := int64(42)
	query, args := BuildRequestLogListQuery(RequestLogListFilter{
		RequestLogID: &requestLogID,
		Limit:        50,
		Offset:       0,
	})

	if !strings.Contains(query, "WHERE id = $1") {
		t.Fatalf("按 request_log_id drill-down 时应使用精确 ID 过滤，SQL=%s", query)
	}
	if !strings.Contains(query, "LIMIT $2 OFFSET $3") {
		t.Fatalf("存在过滤条件时分页参数应顺延，SQL=%s", query)
	}
	if len(args) != 3 || args[0] != requestLogID || args[1] != 50 || args[2] != 0 {
		t.Fatalf("过滤参数顺序不正确: %#v", args)
	}
}

func TestBuildRequestLogListQueryDefaultsToGlobalList(t *testing.T) {
	query, args := BuildRequestLogListQuery(RequestLogListFilter{Limit: 20, Offset: 10})

	if strings.Contains(query, "WHERE") {
		t.Fatalf("未指定过滤条件时应保持全局日志列表，SQL=%s", query)
	}
	if !strings.Contains(query, "LIMIT $1 OFFSET $2") {
		t.Fatalf("无过滤条件时分页参数应从 $1 开始，SQL=%s", query)
	}
	if len(args) != 2 || args[0] != 20 || args[1] != 10 {
		t.Fatalf("分页参数不正确: %#v", args)
	}
}
