package generator

import (
	"context"
	"errors"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"
)

const regularFileMode fs.FileMode = 0o644

func planFiles(outDir string, files []renderedFile) ([]string, []string, error) {
	targets := make([]string, 0, len(files))
	paths := make([]string, 0, len(files))
	for _, file := range files {
		target, fullPath, err := resolveTargetPath(outDir, file.Target)
		if err != nil {
			return nil, nil, err
		}
		targets = append(targets, target)
		paths = append(paths, fullPath)
	}
	return targets, paths, nil
}

func writeFiles(ctx context.Context, outDir string, files []renderedFile, force bool) ([]string, error) {
	targets, paths, err := planFiles(outDir, files)
	if err != nil {
		return nil, err
	}

	if err := prepareOutputDirectory(outDir, paths, force); err != nil {
		return nil, err
	}

	for index, file := range files {
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		fullPath := paths[index]
		if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
			return nil, &FileWriteError{Path: fullPath, Err: err}
		}
		if err := os.WriteFile(fullPath, file.Content, regularFileMode); err != nil {
			return nil, &FileWriteError{Path: fullPath, Err: err}
		}
		if err := os.Chmod(fullPath, regularFileMode); err != nil {
			return nil, &FileWriteError{Path: fullPath, Err: err}
		}
	}

	return targets, nil
}

func prepareOutputDirectory(outDir string, targetPaths []string, force bool) error {
	info, err := os.Stat(outDir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			if err := os.MkdirAll(outDir, 0o755); err != nil {
				return &OutputDirectoryError{Path: outDir, Err: err}
			}
			return nil
		}
		return &OutputDirectoryError{Path: outDir, Err: err}
	}
	if !info.IsDir() {
		return &OutputDirectoryError{Path: outDir, Err: errors.New("path exists and is not a directory")}
	}
	if force {
		return nil
	}

	for _, targetPath := range targetPaths {
		if _, err := os.Stat(targetPath); err == nil {
			return &FileExistsError{Path: targetPath}
		} else if !errors.Is(err, os.ErrNotExist) {
			return &FileWriteError{Path: targetPath, Err: err}
		}
	}

	entries, err := os.ReadDir(outDir)
	if err != nil {
		return &OutputDirectoryError{Path: outDir, Err: err}
	}
	if len(entries) > 0 {
		return &OutputDirectoryNotEmptyError{Path: outDir}
	}
	return nil
}

func resolveTargetPath(outDir string, target string) (string, string, error) {
	cleanTarget, err := cleanTargetPath(target)
	if err != nil {
		return "", "", err
	}

	outputRoot, err := filepath.Abs(outDir)
	if err != nil {
		return "", "", &OutputDirectoryError{Path: outDir, Err: err}
	}
	fullPath := filepath.Join(outputRoot, filepath.FromSlash(cleanTarget))
	relative, err := filepath.Rel(outputRoot, fullPath)
	if err != nil {
		return "", "", &UnsafeTargetPathError{Target: target}
	}
	if relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return "", "", &UnsafeTargetPathError{Target: target}
	}

	return cleanTarget, fullPath, nil
}

func cleanTargetPath(target string) (string, error) {
	trimmed := strings.TrimSpace(target)
	normalized := strings.ReplaceAll(trimmed, "\\", "/")
	if normalized == "" || normalized == "." {
		return "", &UnsafeTargetPathError{Target: target}
	}
	if path.IsAbs(normalized) || filepath.IsAbs(trimmed) || hasWindowsVolumeName(normalized) {
		return "", &UnsafeTargetPathError{Target: target}
	}

	cleaned := path.Clean(normalized)
	if cleaned == "." || cleaned == ".." || strings.HasPrefix(cleaned, "../") {
		return "", &UnsafeTargetPathError{Target: target}
	}
	return cleaned, nil
}

func hasWindowsVolumeName(value string) bool {
	return len(value) >= 2 && value[1] == ':'
}
