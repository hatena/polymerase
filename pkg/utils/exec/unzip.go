package exec

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/pkg/errors"
)

func UnzipIncBackupCmd(ctx context.Context, name, dir string, inc int) error {
	if dir == "" {
		return errors.New("directory path cannot be empty")
	}

	odir := filepath.Join(dir, fmt.Sprintf("inc%d", inc))
	if err := os.MkdirAll(odir, 0755); err != nil {
		return errors.Wrap(err, odir+" dir cannot be created")
	}

	cmd := exec.CommandContext(ctx, "sh", "-c", fmt.Sprintf("gunzip -c %s | xbstream -x -C %s", name, odir))
	if err := cmd.Run(); err != nil {
		return errors.Wrap(err, fmt.Sprintf("failed `gunzip -c %s | xbstream -x -C %s", name, odir))
	}

	if err := os.Remove(name); err != nil {
		return errors.Wrap(err, "failed to remove "+name)
	}

	return nil
}

func UnzipFullBackupCmd(ctx context.Context, name, dir string) error {
	odir := filepath.Join(dir, "base")
	if err := os.MkdirAll(odir, 0755); err != nil {
		return errors.Wrap(err, odir+" dir cannot be created")
	}

	f, err := os.Open(name)
	if err != nil {
		return fmt.Errorf("%s: failed to open archive: %v", name, err)
	}

	gzr, err := gzip.NewReader(f)
	if err != nil {
		return fmt.Errorf("%s: create new gzip reader: %v", name, err)
	}

	if err := untar(tar.NewReader(gzr), odir); err != nil {
		return err
	}

	gzr.Close()
	f.Close()

	if err := os.Remove(name); err != nil {
		return err
	}

	return nil
}

// untar un-tarballs the contents of tr into destination.
func untar(tr *tar.Reader, destination string) error {
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}

		if err := untarFile(tr, header, destination); err != nil {
			return err
		}
	}
	return nil
}

// untarFile untars a single file from tr with header header into destination.
func untarFile(tr *tar.Reader, header *tar.Header, dest string) error {
	switch header.Typeflag {
	case tar.TypeDir:
		return os.MkdirAll(filepath.Join(dest, header.Name), 0755)
	case tar.TypeReg, tar.TypeRegA, tar.TypeChar, tar.TypeBlock, tar.TypeFifo:
		fp := filepath.Join(dest, header.Name)

		if err := os.MkdirAll(filepath.Dir(fp), 0755); err != nil {
			return errors.Wrap(err, "failed to make directory for file")
		}

		out, err := os.Create(fp)
		if err != nil {
			return errors.Wrap(err, "failed to create new file")
		}
		defer out.Close()

		if err := out.Chmod(header.FileInfo().Mode()); err != nil && runtime.GOOS != "windows" {
			return errors.Wrap(err, "failed to change file mode")
		}

		if _, err := io.Copy(out, tr); err != nil {
			return errors.Wrap(err, "failed to write file")
		}

		return nil
	default:
		return fmt.Errorf("%s: unknown type flag: %c", header.Name, header.Typeflag)
	}
}
