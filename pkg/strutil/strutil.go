// Package strutil 字符串工具
package strutil

// DBLen 估算字符串写入 varchar 列(Chinese_PRC_CI_AS，GBK 编码)占用的字节数：
// ASCII 字符占 1 字节，其他字符按 2 字节计算。引擎表的文本列为 varchar，需按字节校验长度。
func DBLen(text string) int {
	n := 0
	for _, r := range text {
		if r < 0x80 {
			n++
		} else {
			n += 2
		}
	}
	return n
}
