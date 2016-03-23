package water

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
	"net/http"
	"path"
	"strings"
)

// from net/http
func cleanPath(p string) string {
	if p == "" {
		return "/"
	}
	if p[0] != '/' {
		p = "/" + p
	}

	// Restful uri
	// path.Clean会去除末尾的'/'("/"除外)
	return path.Clean(p)
}

func requestProxy(req *http.Request) []string {
	if ips := req.Header.Get("X-Forwarded-For"); ips != "" {
		return strings.Split(ips, ",")
	}
	return nil
}

func requestRemoteIp(req *http.Request) string {
	ip := req.Header.Get("X-Real-IP")
	if ip == "" {
		ips := requestProxy(req)
		if len(ips) > 0 && ips[0] != "" {
			return ips[0]
		}
		ip = req.RemoteAddr
		if i := strings.LastIndex(ip, ":"); i > -1 {
			ip = ip[:i]
		}
	}
	return ip
}

// check pattern segment
// 检查url片段的合法性
func checkSplitPattern(pattern string) bool {
	n := len(pattern)
	return n > 0 && pattern[0] == '/' && pattern[n-1] != '/'
}

/*
// wrap http.Handler to middleware
// 将http.Handler包裹成中间件,不推荐
func WrapBefore(handler http.Handler) HandlerFunc {
	return func(ctx *Context) {
		handler.ServeHTTP(ctx.Resp, ctx.Req)

		ctx.Next()
	}
}
*/

// AESEncrypt encrypts text and given key with AES.
func AESEncrypt(key, text []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	b := base64.StdEncoding.EncodeToString(text)
	ciphertext := make([]byte, aes.BlockSize+len(b))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, err
	}
	cfb := cipher.NewCFBEncrypter(block, iv)
	cfb.XORKeyStream(ciphertext[aes.BlockSize:], []byte(b))
	return ciphertext, nil
}

// AESDecrypt decrypts text and given key with AES.
func AESDecrypt(key, text []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	if len(text) < aes.BlockSize {
		return nil, errors.New("ciphertext too short")
	}
	iv := text[:aes.BlockSize]
	text = text[aes.BlockSize:]
	cfb := cipher.NewCFBDecrypter(block, iv)
	cfb.XORKeyStream(text, text)
	data, err := base64.StdEncoding.DecodeString(string(text))
	if err != nil {
		return nil, err
	}
	return data, nil
}
