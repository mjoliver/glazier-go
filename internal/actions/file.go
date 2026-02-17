package actions

import (
	"archive/zip"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/deck"
	"gopkg.in/yaml.v3"
)

// --- file.copy ---

type FileCopyConfig struct {
	Src  string `yaml:"src"`
	Dst  string `yaml:"dst"`
	Dirs bool   `yaml:"dirs"` // copy directories recursively
}

type FileCopy struct{ Config FileCopyConfig }

func NewFileCopy(ctx context.Context, yamlData interface{}) (Action, error) {
	var cfg FileCopyConfig
	data, _ := yaml.Marshal(yamlData)
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("file.copy: %w", err)
	}
	return &FileCopy{Config: cfg}, nil
}

func (a *FileCopy) Validate() error {
	if a.Config.Src == "" || a.Config.Dst == "" {
		return fmt.Errorf("file.copy: src and dst are required")
	}
	return nil
}

func (a *FileCopy) Run(ctx context.Context) error {
	deck.Infof("file.copy: %s -> %s", a.Config.Src, a.Config.Dst)

	info, err := os.Stat(a.Config.Src)
	if err != nil {
		return fmt.Errorf("file.copy: %w", err)
	}

	if info.IsDir() {
		return copyDir(a.Config.Src, a.Config.Dst)
	}
	return copyFile(a.Config.Src, a.Config.Dst)
}

func copyFile(src, dst string) error {
	// Ensure destination directory exists
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}

	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err = io.Copy(out, in); err != nil {
		return err
	}
	return out.Sync()
}

func copyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, _ := filepath.Rel(src, path)
		dstPath := filepath.Join(dst, relPath)

		if info.IsDir() {
			return os.MkdirAll(dstPath, info.Mode())
		}
		return copyFile(path, dstPath)
	})
}

// --- file.mkdir ---

type FileMkdirConfig struct {
	Path string `yaml:"path"`
}

type FileMkdir struct{ Config FileMkdirConfig }

func NewFileMkdir(ctx context.Context, yamlData interface{}) (Action, error) {
	var cfg FileMkdirConfig
	if str, ok := yamlData.(string); ok {
		cfg.Path = str
	} else {
		data, _ := yaml.Marshal(yamlData)
		if err := yaml.Unmarshal(data, &cfg); err != nil {
			return nil, fmt.Errorf("file.mkdir: %w", err)
		}
	}
	return &FileMkdir{Config: cfg}, nil
}

func (a *FileMkdir) Validate() error {
	if a.Config.Path == "" {
		return fmt.Errorf("file.mkdir: path is required")
	}
	return nil
}

func (a *FileMkdir) Run(ctx context.Context) error {
	deck.Infof("file.mkdir: %s", a.Config.Path)
	return os.MkdirAll(a.Config.Path, 0755)
}

// --- file.remove ---

type FileRemoveConfig struct {
	Path string `yaml:"path"`
}

type FileRemove struct{ Config FileRemoveConfig }

func NewFileRemove(ctx context.Context, yamlData interface{}) (Action, error) {
	var cfg FileRemoveConfig
	if str, ok := yamlData.(string); ok {
		cfg.Path = str
	} else {
		data, _ := yaml.Marshal(yamlData)
		if err := yaml.Unmarshal(data, &cfg); err != nil {
			return nil, fmt.Errorf("file.remove: %w", err)
		}
	}
	return &FileRemove{Config: cfg}, nil
}

func (a *FileRemove) Validate() error {
	if a.Config.Path == "" {
		return fmt.Errorf("file.remove: path is required")
	}
	return nil
}

func (a *FileRemove) Run(ctx context.Context) error {
	deck.Infof("file.remove: %s", a.Config.Path)
	return os.RemoveAll(a.Config.Path)
}

// --- file.unzip ---

type FileUnzipConfig struct {
	Src string `yaml:"src"`
	Dst string `yaml:"dst"`
}

type FileUnzip struct{ Config FileUnzipConfig }

func NewFileUnzip(ctx context.Context, yamlData interface{}) (Action, error) {
	var cfg FileUnzipConfig
	data, _ := yaml.Marshal(yamlData)
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("file.unzip: %w", err)
	}
	return &FileUnzip{Config: cfg}, nil
}

func (a *FileUnzip) Validate() error {
	if a.Config.Src == "" || a.Config.Dst == "" {
		return fmt.Errorf("file.unzip: src and dst are required")
	}
	return nil
}

func (a *FileUnzip) Run(ctx context.Context) error {
	deck.Infof("file.unzip: %s -> %s", a.Config.Src, a.Config.Dst)

	r, err := zip.OpenReader(a.Config.Src)
	if err != nil {
		return fmt.Errorf("file.unzip: %w", err)
	}
	defer r.Close()

	for _, f := range r.File {
		fpath := filepath.Join(a.Config.Dst, f.Name)

		// Security: prevent zip slip
		if !strings.HasPrefix(filepath.Clean(fpath), filepath.Clean(a.Config.Dst)+string(os.PathSeparator)) {
			return fmt.Errorf("file.unzip: illegal file path: %s", fpath)
		}

		if f.FileInfo().IsDir() {
			os.MkdirAll(fpath, os.ModePerm)
			continue
		}

		if err := os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
			return err
		}

		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}

		rc, err := f.Open()
		if err != nil {
			outFile.Close()
			return err
		}

		_, err = io.Copy(outFile, rc)
		rc.Close()
		outFile.Close()
		if err != nil {
			return err
		}
	}

	return nil
}

// --- file.download ---

type FileDownloadConfig struct {
	URL string `yaml:"url"`
	Dst string `yaml:"dst"`
}

type FileDownload struct{ Config FileDownloadConfig }

func NewFileDownload(ctx context.Context, yamlData interface{}) (Action, error) {
	var cfg FileDownloadConfig
	data, _ := yaml.Marshal(yamlData)
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("file.download: %w", err)
	}
	return &FileDownload{Config: cfg}, nil
}

func (a *FileDownload) Validate() error {
	if a.Config.URL == "" || a.Config.Dst == "" {
		return fmt.Errorf("file.download: url and dst are required")
	}
	return nil
}

func (a *FileDownload) Run(ctx context.Context) error {
	deck.Infof("file.download: %s -> %s", a.Config.URL, a.Config.Dst)

	client := &http.Client{Timeout: 5 * time.Minute}
	req, err := http.NewRequestWithContext(ctx, "GET", a.Config.URL, nil)
	if err != nil {
		return fmt.Errorf("file.download: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("file.download: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("file.download: bad status: %d", resp.StatusCode)
	}

	// Ensure destination directory
	if err := os.MkdirAll(filepath.Dir(a.Config.Dst), 0755); err != nil {
		return err
	}

	out, err := os.Create(a.Config.Dst)
	if err != nil {
		return fmt.Errorf("file.download: %w", err)
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

// --- Register all file actions ---

func init() {
	Register("file.copy", NewFileCopy)
	Register("file.mkdir", NewFileMkdir)
	Register("file.remove", NewFileRemove)
	Register("file.unzip", NewFileUnzip)
	Register("file.download", NewFileDownload)
}
