package util

import (
	"bytes"
	"context"
	"crypto/aes"
	"crypto/cipher"
	crand "crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"math"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// /////////////////////////////////////////////////////////////////////////////
// BASE62
// /////////////////////////////////////////////////////////////////////////////
// EncodeToBase62 converts a decimal number to base62 string

func EncodeToBase62(num uint64) string {
	if num == 0 {
		return "0"
	}

	const s = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	var buf [15]byte // log62(2^64) ≈ 11.8, 여유 있게 15
	i := len(buf)

	for num > 0 {
		i--
		buf[i] = s[num%62]
		num /= 62
	}

	return string(buf[i:])
}

// DecodeFromBase62 converts a base62 string back to decimal
func DecodeFromBase62(str string) uint64 {
	var result uint64
	for _, char := range str {
		idx := getBase62Index(byte(char))
		temp := result*62 + uint64(idx)
		if temp < result {
			panic(fmt.Errorf("base62 overflow at char: %q", char))
		}
		result = temp
	}
	return result
}

func getBase62Index(char byte) int {
	switch {
	case char >= '0' && char <= '9':
		return int(char - '0')
	case char >= 'A' && char <= 'Z':
		return int(char-'A') + 10
	case char >= 'a' && char <= 'z':
		return int(char-'a') + 36
	default:
		panic(fmt.Errorf("invalid base62 character: %q", char))
	}
}

// /////////////////////////////////////////////////////////////////////////////
// UUID
// /////////////////////////////////////////////////////////////////////////////
/*
UUID는 지극히 낮은 확률로 중복될 수 있음.
단순 태그/임시 식별자 용도로만 사용. 보안적 용도는 UUIDFromCryptoPackage
*/
func UUID() string {
	b := make([]byte, 16)
	rand.New(rand.NewSource(time.Now().UnixNano())).Read(b)
	return fmt.Sprintf("%X", b)
}

func UUIDFromCryptoPackage() string {
	b := make([]byte, 16)
	if _, err := crand.Read(b); err != nil {
		panic(fmt.Errorf("crypto/rand failed: %w", err))
	}
	return UUID4String(b)
}

// UUID4 스펙에 따라 하이픈 삽입한 문자열로 변환
func UUID4String(b []byte) string {
	// UUID version 4 & variant 1 스펙에 따라 비트 조작
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}

// /////////////////////////////////////////////////////////////////////////////
// UKey
// /////////////////////////////////////////////////////////////////////////////
/*
특수용도임
*/
func Ukey(num uint64) string {
	r := uint64(62 + rand.Intn(3843-62+1))
	mul := num * r
	encodedMul := EncodeToBase62(mul)
	encodedR := EncodeToBase62(r)
	return encodedR + encodedMul
}

// /////////////////////////////////////////////////////////////////////////////
// AESGCM
// /////////////////////////////////////////////////////////////////////////////
/*
AES는 16/24/32바이트 사이즈 맞출것
*/
func AESGCMEncrypt(plaintext string, key []byte) (string, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, aesgcm.NonceSize())
	if _, err = io.ReadFull(crand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := aesgcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

/*
AES는 16/24/32바이트 사이즈 맞출것
*/
func AESGCMDecrypt(cipherTextBase64 string, key []byte) (string, error) {
	ciphertext, err := base64.StdEncoding.DecodeString(cipherTextBase64)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := aesgcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return "", fmt.Errorf("ciphertext too short")
	}
	nonce, cipherText := ciphertext[:nonceSize], ciphertext[nonceSize:]

	plainText, err := aesgcm.Open(nil, nonce, cipherText, nil)
	if err != nil {
		return "", err
	}
	return string(plainText), nil
}

// /////////////////////////////////////////////////////////////////////////////
// SHA256
// /////////////////////////////////////////////////////////////////////////////
func SHA256(s string) string {
	h := sha256.New()
	h.Write([]byte(s))
	hashInBytes := h.Sum(nil)
	return hex.EncodeToString(hashInBytes)
}

// 문자열 맨 앞의 '@' At 갯수 1개로 정규화화
func NormalizeAtPrefix(s string) string {
	trimmed := strings.TrimLeftFunc(s, func(r rune) bool {
		return r == '@'
	})
	return "@" + trimmed
}

// 120일간 브라우저가 보관하길 바라지만 브라우저들은 임의 삭제하기도 함.
// 특히 사파리 같은 브라우저들은 단 하루 만에 삭제하기도 함.
func PersistentCookie(name, value string) *http.Cookie {
	return &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Now().Add(120 * 24 * time.Hour), // 120일 후
	}
}

func SessionCookie(name, value string) *http.Cookie {
	return &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	}
}

// ClearCookie는 지정한 이름의 쿠키를 즉시 만료시킵니다.
func ClearCookie(w http.ResponseWriter, name string) {
	cookie := &http.Cookie{
		Name:     name,
		Value:    "",
		Path:     "/",             // 필요에 따라 Path 지정
		Expires:  time.Unix(0, 0), // 1970-01-01로 만료
		MaxAge:   -1,              // 즉시 삭제
		HttpOnly: true,            // 보안 옵션 필요시
		Secure:   true,            // HTTPS만 허용할 경우
	}
	http.SetCookie(w, cookie)
}

// ExePath 는 실행 파일이 위치한 디렉터리 경로를 반환합니다.
// 에러가 발생하면 빈 문자열과 에러를 반환합니다.
func ExePath() (string, error) {
	exePath, err := os.Executable()
	if err != nil {
		return "", err
	}
	return filepath.Dir(exePath), nil
}

// RealPath 는 실행파일 위치 기준 상대 경로를 절대 경로로 변환해 반환합니다.
func RealPath(elem ...string) (string, error) {
	dir, err := ExePath()
	if err != nil {
		return "", err
	}
	paths := append([]string{dir}, elem...)
	return filepath.Join(paths...), nil
}

func Mkdirs(fullPath string) error {
	dirPath := filepath.Dir(fullPath)
	err := os.MkdirAll(dirPath, 0755)
	if err != nil {
		return err
	}
	return nil
}

func Truncate(s string, max int) string {
	runes := []rune(s)
	if len(runes) <= max {
		return s
	}
	return string(runes[:max]) + "..."
}

/*
보안적 무작위성이 필요하다면 crypto/rand 기반으로 교체
*/
func RandNDigits(n int) int {
	if n <= 0 {
		return 0
	}
	min := int(math.Pow10(n - 1))
	max := int(math.Pow10(n)) - 1
	return rand.Intn(max-min+1) + min
}

// HttpPost sends a POST request with given body and headers,
// returns response body as bytes or error.
func HttpPost(
	ctx context.Context, url string, body []byte,
	headers map[string]string,
) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	// 헤더 설정
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// 상태 코드 200~299 아니면 에러 처리
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		data, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(data))
	}

	return io.ReadAll(resp.Body)
}

// DetectAgent 함수: User-Agent 문자열을 받아서 환경/브라우저 구분 반환
func DetectAgent(ua string) string {
	ua = strings.ToLower(ua)

	device := "PC"
	if strings.Contains(ua, "mobile") || strings.Contains(ua, "android") || strings.Contains(ua, "iphone") {
		device = "Mobile"
	} else if strings.Contains(ua, "ipad") {
		device = "Tablet"
	}

	browser := "Unknown"
	switch {
	case strings.Contains(ua, "chrome") && !strings.Contains(ua, "edg"):
		browser = "Chrome"
	case strings.Contains(ua, "safari") && !strings.Contains(ua, "chrome"):
		browser = "Safari"
	case strings.Contains(ua, "firefox"):
		browser = "Firefox"
	case strings.Contains(ua, "edg"):
		browser = "Edge"
	case strings.Contains(ua, "opr") || strings.Contains(ua, "opera"):
		browser = "Opera"
	}

	return fmt.Sprintf("%s %s", device, browser)
}

func Join(args ...string) string {
	return strings.Join(args, ",")
}
