# Package xcrypto 加密工具库

xcrypto 是 oceanho-stack 的统一加解密与哈希工具，覆盖常见安全操作。

## 特性

- **密码哈希**：`BcryptHash` / `BcryptVerify`（bcrypt，推荐存密码）
- **AES**：GCM 认证加密（`AESGCMEncrypt` / `AESGCMDecrypt`），抗篡改
- **RSA**：密钥生成、公钥加密/私钥解密、签名/验签
- **随机**：`RandomString`（可配置长度与字符集）
- **摘要**：`SHA256`、`SHA256Hex`
- **编码**：base64 封装

## 快速开始

```go
// 密码哈希
hash, _ := xcrypto.BcryptHash("secret")
ok := xcrypto.BcryptVerify("secret", hash)

// AES-GCM（key 必须 16/24/32 字节）
ciphertext, _ := xcrypto.AESGCMEncrypt(key32, []byte("data"))
plain, _ := xcrypto.AESGCMDecrypt(key32, ciphertext)

// 随机串
s := xcrypto.RandomString(32)
```

## 安全约定

- 密码存储一律用 bcrypt，禁止明文/普通哈希
- AES 一律用 GCM 认证模式（不要用裸 ECB/CBC）
