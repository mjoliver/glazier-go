package actions

import (
	"archive/zip"
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestFileCopy_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  FileCopyConfig
		wantErr bool
	}{
		{"valid", FileCopyConfig{Src: "a", Dst: "b"}, false},
		{"missing src", FileCopyConfig{Dst: "b"}, true},
		{"missing dst", FileCopyConfig{Src: "a"}, true},
		{"both missing", FileCopyConfig{}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &FileCopy{Config: tt.config}
			err := a.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestFileCopy_Run(t *testing.T) {
	tmp := t.TempDir()

	// Create source file
	srcFile := filepath.Join(tmp, "src.txt")
	os.WriteFile(srcFile, []byte("hello world"), 0644)

	dstFile := filepath.Join(tmp, "dst.txt")

	a := &FileCopy{Config: FileCopyConfig{Src: srcFile, Dst: dstFile}}
	if err := a.Run(context.Background()); err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	data, err := os.ReadFile(dstFile)
	if err != nil {
		t.Fatalf("failed to read destination: %v", err)
	}
	if string(data) != "hello world" {
		t.Errorf("copied content = %q, want %q", string(data), "hello world")
	}
}

func TestFileCopy_RunDir(t *testing.T) {
	tmp := t.TempDir()

	// Create source dir structure
	srcDir := filepath.Join(tmp, "srcdir")
	os.MkdirAll(filepath.Join(srcDir, "sub"), 0755)
	os.WriteFile(filepath.Join(srcDir, "a.txt"), []byte("a"), 0644)
	os.WriteFile(filepath.Join(srcDir, "sub", "b.txt"), []byte("b"), 0644)

	dstDir := filepath.Join(tmp, "dstdir")

	a := &FileCopy{Config: FileCopyConfig{Src: srcDir, Dst: dstDir}}
	if err := a.Run(context.Background()); err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	// Verify files were copied
	data, _ := os.ReadFile(filepath.Join(dstDir, "a.txt"))
	if string(data) != "a" {
		t.Errorf("a.txt = %q, want %q", string(data), "a")
	}

	data, _ = os.ReadFile(filepath.Join(dstDir, "sub", "b.txt"))
	if string(data) != "b" {
		t.Errorf("sub/b.txt = %q, want %q", string(data), "b")
	}
}

func TestFileMkdir_Run(t *testing.T) {
	tmp := t.TempDir()
	dir := filepath.Join(tmp, "deep", "nested", "dir")

	a := &FileMkdir{Config: FileMkdirConfig{Path: dir}}
	if err := a.Run(context.Background()); err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	info, err := os.Stat(dir)
	if err != nil {
		t.Fatalf("Stat() error = %v", err)
	}
	if !info.IsDir() {
		t.Error("expected directory")
	}
}

func TestFileRemove_Run(t *testing.T) {
	tmp := t.TempDir()

	// Create a file to remove
	file := filepath.Join(tmp, "delete_me.txt")
	os.WriteFile(file, []byte("bye"), 0644)

	a := &FileRemove{Config: FileRemoveConfig{Path: file}}
	if err := a.Run(context.Background()); err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	if _, err := os.Stat(file); !os.IsNotExist(err) {
		t.Error("file should have been removed")
	}
}

func TestFileRemove_RunDir(t *testing.T) {
	tmp := t.TempDir()

	dir := filepath.Join(tmp, "rmdir")
	os.MkdirAll(filepath.Join(dir, "sub"), 0755)
	os.WriteFile(filepath.Join(dir, "sub", "file.txt"), []byte("x"), 0644)

	a := &FileRemove{Config: FileRemoveConfig{Path: dir}}
	if err := a.Run(context.Background()); err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	if _, err := os.Stat(dir); !os.IsNotExist(err) {
		t.Error("directory should have been removed")
	}
}

func TestFileUnzip_Run(t *testing.T) {
	tmp := t.TempDir()

	// Create a test zip
	zipPath := filepath.Join(tmp, "test.zip")
	createTestZip(t, zipPath, map[string]string{
		"hello.txt":     "hello",
		"sub/world.txt": "world",
	})

	dstDir := filepath.Join(tmp, "unzipped")
	a := &FileUnzip{Config: FileUnzipConfig{Src: zipPath, Dst: dstDir}}
	if err := a.Run(context.Background()); err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	data, _ := os.ReadFile(filepath.Join(dstDir, "hello.txt"))
	if string(data) != "hello" {
		t.Errorf("hello.txt = %q, want %q", string(data), "hello")
	}

	data, _ = os.ReadFile(filepath.Join(dstDir, "sub", "world.txt"))
	if string(data) != "world" {
		t.Errorf("sub/world.txt = %q, want %q", string(data), "world")
	}
}

func TestFileDownload_Run(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("downloaded content"))
	}))
	defer server.Close()

	tmp := t.TempDir()
	dst := filepath.Join(tmp, "downloaded.txt")

	a := &FileDownload{Config: FileDownloadConfig{URL: server.URL, Dst: dst}}
	if err := a.Run(context.Background()); err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	data, _ := os.ReadFile(dst)
	if string(data) != "downloaded content" {
		t.Errorf("downloaded = %q, want %q", string(data), "downloaded content")
	}
}

func TestFileDownload_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  FileDownloadConfig
		wantErr bool
	}{
		{"valid", FileDownloadConfig{URL: "http://x.com/f", Dst: "/tmp/f"}, false},
		{"missing url", FileDownloadConfig{Dst: "/tmp/f"}, true},
		{"missing dst", FileDownloadConfig{URL: "http://x.com/f"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &FileDownload{Config: tt.config}
			err := a.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// Helper: create a zip file with given files
func createTestZip(t *testing.T, path string, files map[string]string) {
	t.Helper()
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	w := zip.NewWriter(f)
	for name, content := range files {
		fw, err := w.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		fw.Write([]byte(content))
	}
	w.Close()
}
