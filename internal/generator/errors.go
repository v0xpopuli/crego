package generator

import "fmt"

type (
	OutDirRequiredError struct{}

	UnsafeTargetPathError struct {
		Target string
	}

	OutputDirectoryError struct {
		Path string
		Err  error
	}

	OutputDirectoryNotEmptyError struct {
		Path string
	}

	FileExistsError struct {
		Path string
	}

	TemplateRenderError struct {
		Source string
		Err    error
	}

	FileWriteError struct {
		Path string
		Err  error
	}
)

func (e *OutDirRequiredError) Error() string {
	return "generator output directory is required"
}

func (e *UnsafeTargetPathError) Error() string {
	return fmt.Sprintf("unsafe generated file target %q", e.Target)
}

func (e *OutputDirectoryError) Error() string {
	return fmt.Sprintf("prepare output directory %q: %v", e.Path, e.Err)
}

func (e *OutputDirectoryError) Unwrap() error {
	return e.Err
}

func (e *OutputDirectoryNotEmptyError) Error() string {
	return fmt.Sprintf("output directory %q is not empty; use force to overwrite", e.Path)
}

func (e *FileExistsError) Error() string {
	return fmt.Sprintf("target file %q already exists; use force to overwrite", e.Path)
}

func (e *TemplateRenderError) Error() string {
	return fmt.Sprintf("render template %q: %v", e.Source, e.Err)
}

func (e *TemplateRenderError) Unwrap() error {
	return e.Err
}

func (e *FileWriteError) Error() string {
	return fmt.Sprintf("write target file %q: %v", e.Path, e.Err)
}

func (e *FileWriteError) Unwrap() error {
	return e.Err
}
