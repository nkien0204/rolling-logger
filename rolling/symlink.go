package rolling

import (
	"os"
	"path/filepath"
)

func (r *fileLogger) createSymlink() error {
	os.Remove(filepath.Join(r.dir, r.symlinkFileName))
	if err := os.Symlink(r.filename, filepath.Join(r.dir, r.symlinkFileName)); err != nil && !os.IsExist(err) {
		return err
	}
	return nil
}
