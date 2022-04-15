package utils

import (
	"errors"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/mitchellh/go-ps"
	"github.com/neflyte/timetracker/internal/logger"
)

// TrimWithEllipsis trims a string if it is longer than trimLength and appends an ellipsis if the string was trimmed
func TrimWithEllipsis(toTrim string, trimLength int) string {
	if len(toTrim) <= trimLength {
		return toTrim
	}
	return toTrim[0:trimLength-2] + `…`
}

// CheckPidfile reads a PID from the specified file and verifies if a process with that PID is running. Returns
// os.ErrNotExist if the pidfile does not exist, otherwise returns an error wrapped in utils.ErrCheckPidfile if an
// error occurred during the check or utils.ErrStalePidfile if the PID file is stale.
func CheckPidfile(pidFilename string) (bool, error) {
	log := logger.GetLogger("CheckPidfile")
	if pidFilename == "" {
		return false, errors.New("empty pidfile name")
	}
	pidfile, err := os.Open(pidFilename)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return false, err
		}
		return false, ErrCheckPidfile{err}
	}
	defer func() {
		err = pidfile.Close()
		if err != nil {
			log.Err(err).Msg("error closing pidfile")
		}
	}()
	pidfileContents, err := io.ReadAll(pidfile)
	if err != nil {
		return false, ErrCheckPidfile{err}
	}
	cleaned := strings.TrimSuffix(string(pidfileContents), "\n")
	pid, err := strconv.Atoi(cleaned)
	if err != nil {
		return false, ErrCheckPidfile{err}
	}
	proc, err := ps.FindProcess(pid)
	if err != nil {
		return false, ErrCheckPidfile{err}
	}
	if proc == nil {
		return false, ErrStalePidfile
	}
	return true, nil
}
