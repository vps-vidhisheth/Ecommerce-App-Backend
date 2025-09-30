package util

import (
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func SaveUploadedFile(file multipart.File, handler *multipart.FileHeader, uploadDir string, fullName string) (string, error) {
	ext := strings.ToLower(filepath.Ext(handler.Filename))
	allowedExt := map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".xlsx": true,
		".xls":  true,
	}

	if !allowedExt[ext] {
		return "", errors.New("invalid file type: only .jpg, .jpeg, .png allowed")
	}

	err := os.MkdirAll(uploadDir, os.ModePerm) // mode perm -the permissions for the new directories  value is 0777 in octal  - read/ write /execute
	if err != nil {
		return "", err
	}

	filename := fmt.Sprintf("%s_%d%s", fullName, time.Now().UnixNano(), ext)
	filepath := filepath.Join(uploadDir, filename)

	dst, err := os.Create(filepath)
	if err != nil {
		return "", err
	}
	defer dst.Close()

	_, err = io.Copy(dst, file)
	if err != nil {
		return "", err
	}

	return filename, nil
}
