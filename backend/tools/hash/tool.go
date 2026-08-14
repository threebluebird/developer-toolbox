package hashtool

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"developer-toolbox/backend/models"
	"encoding/hex"
	"errors"
	"hash"
	"io"
	"os"
	"strings"
)

type HashTool struct{}

func NewHashTool() *HashTool {
	return &HashTool{}
}

func (t *HashTool) Info() models.Tool {
	return models.Tool{
		ID:           "hash",
		Name:         "Hash",
		Description:  "Compute text hashes using MD5, SHA1, SHA256, and SHA512.",
		Category:     "encoding",
		Icon:         "hash",
		Version:      "0.1.0",
		Capabilities: models.ToolCapabilities{FileInput: true, Streaming: true},
		Keywords:     []string{"hash", "checksum", "md5", "sha1", "sha256", "sha512", "file"},
	}
}

func (t *HashTool) Execute(input models.ToolInput) models.ToolOutput {
	payload := input.Payload
	if payload == nil {
		return errorResponse("payload required")
	}

	algorithm, _ := payload["algorithm"].(string)
	algorithm = strings.ToLower(strings.TrimSpace(algorithm))
	if algorithm == "" {
		algorithm = "sha256"
	}

	var hashValue string
	var err error
	// filePath 优先于文本输入：文件通过流式 io.Copy 计算，避免大文件整体载入内存。
	if filePath, ok := payload["filePath"].(string); ok && strings.TrimSpace(filePath) != "" {
		hashValue, err = computeFileHash(algorithm, filePath)
	} else {
		raw, ok := payload["input"].(string)
		if !ok {
			return errorResponse("input must be a string")
		}
		hashValue, err = computeHash(algorithm, raw)
	}
	if err != nil {
		return errorResponse(err.Error())
	}

	return models.ToolOutput{Success: true, Data: hashValue}
}

func computeHash(algorithm, input string) (string, error) {
	switch algorithm {
	case "md5":
		h := md5.Sum([]byte(input))
		return hex.EncodeToString(h[:]), nil
	case "sha1":
		h := sha1.Sum([]byte(input))
		return hex.EncodeToString(h[:]), nil
	case "sha256":
		h := sha256.Sum256([]byte(input))
		return hex.EncodeToString(h[:]), nil
	case "sha512":
		h := sha512.Sum512([]byte(input))
		return hex.EncodeToString(h[:]), nil
	default:
		return "", errors.New("unsupported algorithm")
	}
}

func errorResponse(message string) models.ToolOutput {
	return models.Failure("INVALID_HASH", message)
}

func newHasher(algorithm string) (hash.Hash, error) {
	switch algorithm {
	case "md5":
		return md5.New(), nil
	case "sha1":
		return sha1.New(), nil
	case "sha256":
		return sha256.New(), nil
	case "sha512":
		return sha512.New(), nil
	default:
		return nil, errors.New("unsupported algorithm")
	}
}

func computeFileHash(algorithm, filePath string) (string, error) {
	// 先创建算法实例再打开文件，可在算法非法时避免不必要的文件句柄操作。
	hasher, err := newHasher(algorithm)
	if err != nil {
		return "", err
	}
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()
	if _, err := io.Copy(hasher, file); err != nil {
		return "", err
	}
	return hex.EncodeToString(hasher.Sum(nil)), nil
}
