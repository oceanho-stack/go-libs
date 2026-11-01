package xcrypto

import (
	"bytes"
	"testing"
)

func TestBcrypt(t *testing.T) {
	hash, err := BcryptHash("my-password")
	if err != nil {
		t.Fatal(err)
	}
	if hash == "my-password" {
		t.Error("哈希不应等于明文")
	}
	if !BcryptVerify("my-password", hash) {
		t.Error("正确密码应验证通过")
	}
	if BcryptVerify("wrong", hash) {
		t.Error("错误密码不应验证通过")
	}
}

func TestAESGCMRoundTrip(t *testing.T) {
	key := bytes.Repeat([]byte{0x42}, 32) // 32 字节 key
	plain := []byte("sensitive data 数据")

	enc, err := AESGCMEncrypt(key, plain)
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Equal(enc, plain) {
		t.Error("密文不应等于明文")
	}

	dec, err := AESGCMDecrypt(key, enc)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(dec, plain) {
		t.Errorf("解密结果不一致: %q", dec)
	}
}

func TestAESGCMTamper(t *testing.T) {
	key := bytes.Repeat([]byte{0x11}, 16)
	plain := []byte("data")
	enc, _ := AESGCMEncrypt(key, plain)

	// 篡改密文一个字节，解密应失败
	enc[len(enc)-1] ^= 0xff
	if _, err := AESGCMDecrypt(key, enc); err == nil {
		t.Error("篡改密文应解密失败")
	}
}

func TestAESGCMInvalidKey(t *testing.T) {
	if _, err := AESGCMEncrypt([]byte("short"), []byte("x")); err == nil {
		t.Error("非16/24/32字节key应报错")
	}
}

func TestRandomString(t *testing.T) {
	s, err := RandomString(32)
	if err != nil {
		t.Fatal(err)
	}
	if len(s) != 32 {
		t.Errorf("len = %d, want 32", len(s))
	}
	if s == "" {
		t.Error("随机串为空")
	}
	// 两次不同
	s2, _ := RandomString(32)
	if s == s2 {
		t.Error("两次随机串相同")
	}
	// 空长度
	if s, _ := RandomString(0); s != "" {
		t.Errorf("RandomString(0) = %q, want 空", s)
	}
}

func TestSHA256(t *testing.T) {
	// 已知向量：sha256("abc")
	want := "ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad"
	if got := SHA256Hex([]byte("abc")); got != want {
		t.Errorf("sha256(abc) = %s, want %s", got, want)
	}
	if len(SHA256([]byte("x"))) != 32 {
		t.Error("sha256 应为 32 字节")
	}
}

func TestBase64(t *testing.T) {
	orig := []byte("hello 世界")
	enc := EncodeBase64(orig)
	dec, err := DecodeBase64(enc)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(dec, orig) {
		t.Error("base64 往返不一致")
	}
}

func TestRSAKeyGenAndPEM(t *testing.T) {
	priv, err := GenerateRSAKey(2048)
	if err != nil {
		t.Fatal(err)
	}

	// PEM 导出
	privPEM := EncodeRSAPrivateKeyPEM(priv)
	pubPEM := EncodeRSAPublicKeyPEM(&priv.PublicKey)
	if len(privPEM) == 0 || len(pubPEM) == 0 {
		t.Fatal("PEM 导出为空")
	}

	// 解析回
	parsedPriv, err := ParseRSAPrivateKeyPEM(privPEM)
	if err != nil {
		t.Fatalf("解析私钥失败: %v", err)
	}
	if parsedPriv.D.Cmp(priv.D) != 0 {
		t.Error("私钥解析不一致")
	}

	parsedPub, err := ParseRSAPublicKeyPEM(pubPEM)
	if err != nil {
		t.Fatalf("解析公钥失败: %v", err)
	}
	if parsedPub.N.Cmp(priv.N) != 0 {
		t.Error("公钥解析不一致")
	}
}
