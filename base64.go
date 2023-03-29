package mail

import "bytes"

func base64LineBreaker(s []byte) []byte {
	var buf bytes.Buffer
	for len(s) > 76 {
		buf.Write(s[:76])
		buf.WriteString("\r\n")
		s = s[76:]
	}
	buf.Write(s)
	return buf.Bytes()
}
