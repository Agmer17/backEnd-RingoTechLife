package storage

import (
	"context"
	"errors"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/google/uuid"
	"golang.org/x/sync/errgroup"
)

var allowedMimes = map[string]bool{
	// image
	"image/jpeg": true,
	"image/png":  true,
	"image/webp": true,

	// video
	"video/mp4":       true,
	"video/webm":      true,
	"video/quicktime": true,
}

const (
	TypeVideo    = "VIDEO"
	TypeAudio    = "AUDIO"
	TypeImage    = "IMAGE"
	TypeDocument = "DOCUMENT"
)

type FileStorage struct {
	Root   string
	Public string
}

func mustGetProjectRoot() string {
	exe, err := os.Executable()
	if err != nil {
		panic(err)
	}
	return filepath.Dir(exe)
}

func mustCreateDir(path string) {
	if err := os.MkdirAll(path, 0755); err != nil {
		panic(err)
	}
}

func NewFileServerStorage() *FileStorage {

	root := mustGetProjectRoot()
	uploadsDir := filepath.Join(root, "uploads")

	publicDir := filepath.Join(uploadsDir, "public")

	mustCreateDir(uploadsDir) // uploads
	mustCreateDir(publicDir)  // public

	fileStoreage := FileStorage{
		Root:   root,
		Public: publicDir,
	}

	return &fileStoreage

}

func (storage *FileStorage) SavePublicFile(
	fileHeader *multipart.FileHeader,
	ext string,
	place ...string) (string, error) {

	file, err := fileHeader.Open()
	if err != nil {
		return "", err
	}
	defer file.Close()

	fileName := uuid.New().String() + ext

	fullPath := storage.GetPathPublicFile(fileName, place...)

	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		return "", err
	}

	dst, err := os.Create(fullPath)
	if err != nil {
		return "", err
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		return "", err
	}

	return fileName, nil

}

func (storage *FileStorage) DetectFileType(fileHeader *multipart.FileHeader) (string, error) {

	f, err := fileHeader.Open()

	if err != nil {
		return "", err
	}

	defer f.Close()

	buf := make([]byte, 512)
	_, err = f.Read(buf)

	if err != nil {
		return "", err
	}

	mimeType := http.DetectContentType(buf)

	return mimeType, nil

}

func (storage *FileStorage) IsTypeSupportted(mimeType string) (string, bool) {

	if !allowedMimes[mimeType] {
		return "", false
	}

	_, after, _ := strings.Cut(mimeType, "/")

	fileExt := "." + after

	return fileExt, true
}

func (storage *FileStorage) GetMediaType(mime string) string {
	switch {
	case strings.HasPrefix(mime, "image/"):
		return TypeImage
	case strings.HasPrefix(mime, "video/"):
		return TypeVideo
	default:
		return ""
	}
}

func (storage *FileStorage) SaveAllPublicFiles(
	context context.Context,
	files []*multipart.FileHeader,
	filesExt []string,
	place ...string) ([]string, error) {

	if len(files) != len(filesExt) {
		return []string{}, errors.New("The files len and files ext len aren't the same!")
	}

	results := make([]string, len(files))

	tpool, ctx := errgroup.WithContext(context)

	tpool.SetLimit(runtime.NumCPU())

	// for loop save file multithread
	for index, file := range files {

		tpool.Go(func() error {

			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
			}

			filename, err := storage.SavePublicFile(file, filesExt[index], place...)

			if err != nil {
				return err
			}

			results[index] = filename

			return nil
		})

	}

	if err := tpool.Wait(); err != nil {
		return results, err
	}

	return results, nil
}

func (storage *FileStorage) DeletePublicFile(fname string, place ...string) {

	parts := []string{storage.Public}
	parts = append(parts, place...)
	parts = append(parts, fname)

	deletePath := filepath.Join(parts...)

	if err := os.Remove(deletePath); err != nil {
		if !os.IsNotExist(err) {
			log.Printf("failed to remove file %s: %v", deletePath, err)
		}
	}

}

func (storage *FileStorage) DeleteAllPublicFile(fileNames []string, place ...string) {
	for _, v := range fileNames {
		storage.DeletePublicFile(v, place...)
	}
}

func (storage *FileStorage) GetPathPublicFile(filename string, place ...string) string {
	parts := []string{
		storage.Public,
	}
	parts = append(parts, place...)
	parts = append(parts, filename)
	return path.Join(parts...)
}
