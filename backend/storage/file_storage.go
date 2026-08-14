package storage

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
)

type FileStorage struct {
	basePath string
}

func NewFileStorage(basePath string) *FileStorage {
	return &FileStorage{basePath: basePath}
}

func (s *FileStorage) Get(key string) ([]byte, error) {
	path, err := s.pathForKey(key)
	if err != nil {
		return nil, err
	}
	return os.ReadFile(path)
}

func (s *FileStorage) Set(key string, data []byte) error {
	// 延迟创建目录，使纯读取场景不会无意义地修改文件系统。
	if err := os.MkdirAll(s.basePath, 0o755); err != nil {
		return err
	}
	path, err := s.pathForKey(key)
	if err != nil {
		return err
	}
	// Write and sync a sibling temporary file before replacing the destination.
	// This prevents a crash during persistence from leaving truncated JSON.
	temporary, err := os.CreateTemp(s.basePath, ".toolbox-*.tmp")
	if err != nil {
		return err
	}
	temporaryPath := temporary.Name()
	defer os.Remove(temporaryPath)
	if err := temporary.Chmod(0o600); err != nil {
		_ = temporary.Close()
		return err
	}
	if _, err := temporary.Write(data); err != nil {
		_ = temporary.Close()
		return err
	}
	if err := temporary.Sync(); err != nil {
		_ = temporary.Close()
		return err
	}
	if err := temporary.Close(); err != nil {
		return err
	}
	// Windows cannot atomically replace an existing path with os.Rename. The
	// fallback window is short and the complete temp file remains recoverable.
	if err := os.Rename(temporaryPath, path); err == nil {
		return nil
	}
	if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return os.Rename(temporaryPath, path)
}

func (s *FileStorage) Delete(key string) error {
	path, err := s.pathForKey(key)
	if err != nil {
		return err
	}
	return os.Remove(path)
}

func (s *FileStorage) pathForKey(key string) (string, error) {
	// key 只能是当前目录下的文件名；拒绝绝对路径和路径分隔符，
	// 防止 Repository 写入 basePath 之外的位置。
	if key == "" || filepath.IsAbs(key) || strings.ContainsAny(key, `/\\`) || key == "." || key == ".." {
		return "", errors.New("invalid storage key")
	}
	return filepath.Join(s.basePath, key), nil
}
