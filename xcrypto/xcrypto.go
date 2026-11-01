// Package xcrypto 统一加解密与哈希工具：密码哈希/AES-GCM/RSA/随机/摘要/编码。
package xcrypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/pem"
	"errors"
	"fmt"
	"math/big"

	"golang.org/x/crypto/bcrypt"
)

// ---------- 密码哈希（bcrypt） ----------

// BcryptHash 对密码做 bcrypt 哈希
func BcryptHash(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// BcryptVerify 校验密码与哈希是否匹配
func BcryptVerify(password, hash string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}

// ---------- AES-GCM ----------

// AESGCMEncrypt AES-GCM 认证加密。key 长度必须为 16/24/32 字节。
// 返回格式：nonce(12字节) + ciphertext，自带认证标签。
func AESGCMEncrypt(key, plaintext []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil, err
	}
	sealed := gcm.Seal(nonce, nonce, plaintext, nil)
	return sealed, nil
}

// AESGCMDecrypt AES-GCM 解密，校验完整性。输入为 AESGCMEncrypt 的输出。
func AESGCMDecrypt(key, ciphertext []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	if len(ciphertext) < gcm.NonceSize() {
		return nil, errors.New("xcrypto: ciphertext 过短")
	}
	nonce := ciphertext[:gcm.NonceSize()]
	sealed := ciphertext[gcm.NonceSize():]
	return gcm.Open(nil, nonce, sealed, nil)
}

// ---------- 随机串 ----------

const (
	lowerChars   = "abcdefghijklmnopqrstuvwxyz"
	upperChars   = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	digitChars   = "0123456789"
	alphaNumChars = lowerChars + upperChars + digitChars
)

// RandomString 生成指定长度的随机字符串（字母+数字）。
// n<=0 返回空串。
func RandomString(n int) (string, error) {
	if n <= 0 {
		return "", nil
	}
	return randomFrom(n, alphaNumChars)
}

func randomFrom(n int, charset string) (string, error) {
	b := make([]byte, n)
	maxIdx := big.NewInt(int64(len(charset)))
	for i := 0; i < n; i++ {
		idx, err := rand.Int(rand.Reader, maxIdx)
		if err != nil {
			return "", err
		}
		b[i] = charset[idx.Int64()]
	}
	return string(b), nil
}

// ---------- 摘要 ----------

// SHA256 计算 sha256 摘要（原始字节）
func SHA256(data []byte) []byte {
	h := sha256.Sum256(data)
	return h[:]
}

// SHA256Hex 计算 sha256 摘要（hex 字符串）
func SHA256Hex(data []byte) string {
	return hex.EncodeToString(SHA256(data))
}

// ---------- 编码 ----------

// EncodeBase64 base64 编码
func EncodeBase64(data []byte) string {
	return base64.StdEncoding.EncodeToString(data)
}

// DecodeBase64 base64 解码
func DecodeBase64(s string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(s)
}

// ---------- RSA ----------

// GenerateRSAKey 生成 RSA 密钥对（bits 建议 2048）
func GenerateRSAKey(bits int) (*rsa.PrivateKey, error) {
	if bits == 0 {
		bits = 2048
	}
	return rsa.GenerateKey(rand.Reader, bits)
}

// EncodeRSAPublicKeyPEM 导出公钥为 PEM（PKIX）
func EncodeRSAPublicKeyPEM(pub *rsa.PublicKey) []byte {
	der, err := x509.MarshalPKIXPublicKey(pub)
	if err != nil {
		return nil
	}
	return pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: der})
}

// EncodeRSAPrivateKeyPEM 导出私钥为 PEM（PKCS1）
func EncodeRSAPrivateKeyPEM(priv *rsa.PrivateKey) []byte {
	return pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(priv),
	})
}

// ParseRSAPrivateKeyPEM 解析 PEM 私钥
func ParseRSAPrivateKeyPEM(pemBytes []byte) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode(pemBytes)
	if block == nil {
		return nil, errors.New("xcrypto: 无法解析 PEM")
	}
	return x509.ParsePKCS1PrivateKey(block.Bytes)
}

// ParseRSAPublicKeyPEM 解析 PEM 公钥
func ParseRSAPublicKeyPEM(pemBytes []byte) (*rsa.PublicKey, error) {
	block, _ := pem.Decode(pemBytes)
	if block == nil {
		return nil, errors.New("xcrypto: 无法解析 PEM")
	}
	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	p, ok := pub.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("xcrypto: 非 RSA 公钥")
	}
	return p, nil
}
