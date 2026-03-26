package internal

import (
	"fmt"
	"mime/multipart"
	"strconv"
	"strings"
)

var paramsMaxLen = 2048

// SafeKey 以下参数不会以明文保存到日志中，仅保留参数名
var SafeKey = map[string]bool{
	`password`:        true,
	`old_password`:    true,
	`secret`:          true,
	`token`:           true,
	`xy_token`:        true,
	`xy_device_token`: true,
	`sign`:            true,
	`_sign`:           true,
}

// ToLogValue 日志数据，会根据k决定是否隐藏日志，返回string或[]string
func ToLogValue(k string, v any, desensitization bool) any {
	if desensitization && SafeKey[k] {
		v = ToSafeValue(v)
	}
	switch val := v.(type) {
	case *multipart.FileHeader:
		return ToLogFileValue(val)
	case []*multipart.FileHeader:
		ls := v.([]*multipart.FileHeader)
		l := len(ls)
		if l == 1 {
			return ToLogFileValue(ls[0])
		}
		list := make([]string, 0, l)
		for _, vv := range ls {
			list = append(list, ToLogFileValue(vv))
		}
		return list
	case string:
		return ToLogStringValue(v.(string))
	case []string:
		ls := v.([]string)
		l := len(ls)
		if l == 1 {
			return ToLogStringValue(ls[0])
		}
		list := make([]string, 0, l)
		for _, vv := range ls {
			list = append(list, ToLogStringValue(vv))
		}
		return list
	}
	return v
}

// ToLogStringValue 日志字符串数据，如果过长会被截断
func ToLogStringValue(v string, maxLen ...int) string {
	size := len(v)
	if len(maxLen) != 0 && maxLen[0] != 0 {
		if size > maxLen[0] {
			return strings.Join([]string{v[:maxLen[0]], `...(`, strconv.Itoa(size), `)`}, ``)
		}
	} else if size > paramsMaxLen {
		return strings.Join([]string{v[:paramsMaxLen], `...(`, strconv.Itoa(size), `)`}, ``)
	}
	return v
}

// ToLogFileValue 日志文件数据，会将文件摘要拼成字符串
func ToLogFileValue(v *multipart.FileHeader) string {
	return fmt.Sprintf(`file(%s) name: %s size: %d`, v.Header.Get(`content-type`), v.Filename, v.Size)
}

const maskChar = "-" // 使用"-"而不是"*"是因为放到query时*会被转换为%2A不方便查看
const maskChar2 = "---safe---"

func ToSafeString(s string) string {
	return toSafeString(s)
}
func ToSafeValue(v any) any {
	if s, ok := v.(string); ok {
		return toSafeString(s)
	} else if vv, ok := v.([]string); ok {
		vvv := make([]string, len(vv))
		for i, s := range vv {
			vvv[i] = toSafeString(s)
		}
		return vvv
	} else {
		return maskChar2
	}
}

func toSafeString(s string) string {
	l := len(s)
	if l < 4 {
		return strings.Repeat(maskChar, l)
	}
	masked := l >> 1            // 总长度除2
	prefix := (l - masked) >> 1 // 剩余字符数除2
	suffix := l - masked - prefix
	return s[:prefix] + strings.Repeat(maskChar, masked) + s[l-suffix:]
}
