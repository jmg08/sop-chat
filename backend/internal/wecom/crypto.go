package wecom

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha1"
	"encoding/base64"
	"encoding/binary"
	"encoding/xml"
	"errors"
	"fmt"
	"math/rand"
	"sort"
	"strings"
	"time"
)

// WXBizMsgCrypt 企业微信消息加解密
type WXBizMsgCrypt struct {
	token          string
	encodingAESKey string
	corpID         string
	aesKey         []byte
}

// NewWXBizMsgCrypt 创建消息加解密实例
func NewWXBizMsgCrypt(token, encodingAESKey, corpID string) (*WXBizMsgCrypt, error) {
	if len(encodingAESKey) != 43 {
		return nil, errors.New("encodingAESKey 长度必须为 43 位")
	}

	aesKey, err := base64.StdEncoding.DecodeString(encodingAESKey + "=")
	if err != nil {
		return nil, fmt.Errorf("解码 encodingAESKey 失败: %w", err)
	}

	return &WXBizMsgCrypt{
		token:          token,
		encodingAESKey: encodingAESKey,
		corpID:         corpID,
		aesKey:         aesKey,
	}, nil
}

// VerifyURL 验证回调 URL 有效性
func (c *WXBizMsgCrypt) VerifyURL(msgSignature, timestamp, nonce, echoStr string) (string, error) {
	signature := c.calcSignature(timestamp, nonce, echoStr)
	if signature != msgSignature {
		return "", errors.New("签名验证失败")
	}

	plaintext, err := c.decrypt(echoStr)
	if err != nil {
		return "", fmt.Errorf("解密 echostr 失败: %w", err)
	}

	return plaintext, nil
}

// DecryptMsg 解密消息
func (c *WXBizMsgCrypt) DecryptMsg(msgSignature, timestamp, nonce string, postData []byte) ([]byte, error) {
	var encryptedMsg struct {
		ToUserName string `xml:"ToUserName"`
		AgentID    string `xml:"AgentID"`
		Encrypt    string `xml:"Encrypt"`
	}

	if err := xml.Unmarshal(postData, &encryptedMsg); err != nil {
		return nil, fmt.Errorf("解析加密消息 XML 失败: %w", err)
	}

	signature := c.calcSignature(timestamp, nonce, encryptedMsg.Encrypt)
	if signature != msgSignature {
		return nil, errors.New("消息签名验证失败")
	}

	plaintext, err := c.decrypt(encryptedMsg.Encrypt)
	if err != nil {
		return nil, fmt.Errorf("解密消息失败: %w", err)
	}

	return []byte(plaintext), nil
}

// EncryptMsg 加密消息
func (c *WXBizMsgCrypt) EncryptMsg(replyMsg, timestamp, nonce string) ([]byte, error) {
	encrypted, err := c.encrypt(replyMsg)
	if err != nil {
		return nil, fmt.Errorf("加密消息失败: %w", err)
	}

	signature := c.calcSignature(timestamp, nonce, encrypted)

	response := fmt.Sprintf(`<xml>
<Encrypt><![CDATA[%s]]></Encrypt>
<MsgSignature><![CDATA[%s]]></MsgSignature>
<TimeStamp>%s</TimeStamp>
<Nonce><![CDATA[%s]]></Nonce>
</xml>`, encrypted, signature, timestamp, nonce)

	return []byte(response), nil
}

// calcSignature 计算签名
func (c *WXBizMsgCrypt) calcSignature(timestamp, nonce, encrypt string) string {
	strs := []string{c.token, timestamp, nonce, encrypt}
	sort.Strings(strs)
	str := strings.Join(strs, "")

	hash := sha1.New()
	hash.Write([]byte(str))
	return fmt.Sprintf("%x", hash.Sum(nil))
}

// decrypt 解密
func (c *WXBizMsgCrypt) decrypt(encrypted string) (string, error) {
	ciphertext, err := base64.StdEncoding.DecodeString(encrypted)
	if err != nil {
		return "", fmt.Errorf("base64 解码失败: %w", err)
	}

	block, err := aes.NewCipher(c.aesKey)
	if err != nil {
		return "", fmt.Errorf("创建 AES cipher 失败: %w", err)
	}

	if len(ciphertext) < aes.BlockSize {
		return "", errors.New("密文长度不足")
	}

	iv := c.aesKey[:aes.BlockSize]
	mode := cipher.NewCBCDecrypter(block, iv)

	plaintext := make([]byte, len(ciphertext))
	mode.CryptBlocks(plaintext, ciphertext)

	plaintext = pkcs7Unpad(plaintext)
	if len(plaintext) < 20 {
		return "", errors.New("解密后数据长度不足")
	}

	content := plaintext[16:]
	msgLen := binary.BigEndian.Uint32(content[:4])
	if int(msgLen) > len(content)-4 {
		return "", errors.New("消息长度字段无效")
	}

	msg := content[4 : 4+msgLen]
	receivedCorpID := string(content[4+msgLen:])

	if receivedCorpID != c.corpID {
		return "", fmt.Errorf("corpID 不匹配: 期望 %s, 收到 %s", c.corpID, receivedCorpID)
	}

	return string(msg), nil
}

// encrypt 加密
func (c *WXBizMsgCrypt) encrypt(plaintext string) (string, error) {
	randomBytes := make([]byte, 16)
	rand.New(rand.NewSource(time.Now().UnixNano())).Read(randomBytes) //nolint:gosec

	msgLenBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(msgLenBytes, uint32(len(plaintext)))

	var buf bytes.Buffer
	buf.Write(randomBytes)
	buf.Write(msgLenBytes)
	buf.WriteString(plaintext)
	buf.WriteString(c.corpID)

	padded := pkcs7Pad(buf.Bytes())

	block, err := aes.NewCipher(c.aesKey)
	if err != nil {
		return "", fmt.Errorf("创建 AES cipher 失败: %w", err)
	}

	iv := c.aesKey[:aes.BlockSize]
	mode := cipher.NewCBCEncrypter(block, iv)

	ciphertext := make([]byte, len(padded))
	mode.CryptBlocks(ciphertext, padded)

	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// pkcs7Pad PKCS7 填充
func pkcs7Pad(data []byte) []byte {
	blockSize := aes.BlockSize
	padding := blockSize - len(data)%blockSize
	padText := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(data, padText...)
}

// pkcs7Unpad PKCS7 去填充（企业微信使用 32 字节块大小）
func pkcs7Unpad(data []byte) []byte {
	if len(data) == 0 {
		return data
	}
	padding := int(data[len(data)-1])
	if padding < 1 || padding > 32 {
		return data
	}
	if padding > len(data) {
		return data
	}
	for i := len(data) - padding; i < len(data); i++ {
		if data[i] != byte(padding) {
			return data
		}
	}
	return data[:len(data)-padding]
}
