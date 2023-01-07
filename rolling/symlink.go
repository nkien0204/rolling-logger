package rolling

import (
	"os"
	"path/filepath"
)

func (r *rolling) createSymlink() error {
	if err := os.Symlink(r.filename, filepath.Join(r.dir, r.symlinkFileName)); err != nil {
		return err
	}
	return nil
}
