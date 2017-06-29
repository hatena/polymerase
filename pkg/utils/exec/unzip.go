package exec

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

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

	if err := exec.CommandContext(ctx, "tar", "xf", name, "-C", odir).Run(); err != nil {
		return errors.Wrap(err, "Failed unzip")
	}

	if err := os.Remove(name); err != nil {
		return err
	}

	return nil
}
