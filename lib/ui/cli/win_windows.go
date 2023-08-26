package cli

/*
This file was adapted from:
https://github.com/safing/portmaster/blob/develop/cmds/portmaster-start/console_windows.go
*/

import (
	"os"
	"syscall"

	"github.com/neflyte/timetracker/lib/logger"
	"golang.org/x/sys/windows"
)

const (
	windowsAttachParentProcess = ^uintptr(0) // (DWORD)-1
)

var (
	kernel32          = syscall.NewLazyDLL("kernel32.dll")
	procAttachConsole = kernel32.NewProc("AttachConsole")
)

// AttachToParentConsole
//
// Windows console output is a mess.
//
// If you compile as "-H windows", then if you launch your program without
// a console, Windows forcibly creates one to use as your stdin/stdout, which
// is silly for a GUI app, so we can't do that.
//
// If you compile as "-H windowsgui", then it doesn't create a console for
// your app... but also doesn't provide a working stdin/stdout/stderr even if
// you *did* launch from the console.  However, you can use AttachConsole()
// to get a handle to your parent process's console, if any, and then
// os.NewFile() to turn that handle into a fd usable as stdout/stderr.
//
// However, then you have the problem that if you redirect stdout or stderr
// from the shell, you end up ignoring the redirection by forcing it to the
// console.
//
// To fix *that*, we have to detect whether there was a pre-existing stdout
// or not. We can check GetStdHandle(), which returns 0 for "should be
// console" and nonzero for "already pointing at a file."
//
// Be careful though!  As soon as you run AttachConsole(), it resets *all*
// the GetStdHandle() handles to point them at the console instead, thus
// throwing away the original file redirects.  So we have to GetStdHandle()
// *before* AttachConsole().
//
// For some reason, powershell redirections provide a valid file handle, but
// writing to that handle doesn't write to the file.  I haven't found a way
// to work around that.  (Windows 10.0.17763.379)
//
// Net result is as follows.
// Before:
//
//	SHELL            NON-REDIRECTED     REDIRECTED
//	explorer.exe     no console         n/a
//	cmd.exe          broken             works
//	powershell       broken             broken
//	WSL bash         broken             works
//
// After
//
//	SHELL            NON-REDIRECTED     REDIRECTED
//	explorer.exe     no console         n/a
//	cmd.exe          works              works
//	powershell       works              broken
//	WSL bash         works              works
//
// We don't seem to make anything worse, at least.
// nolint:cyclop
func AttachToParentConsole() (bool, error) {
	log := logger.GetFuncLogger(packageLogger, "AttachToParentConsole")
	// get std handles before we attempt to attach to parent console
	stdin, _ := syscall.GetStdHandle(syscall.STD_INPUT_HANDLE)   // nolint:errcheck
	stdout, _ := syscall.GetStdHandle(syscall.STD_OUTPUT_HANDLE) // nolint:errcheck
	stderr, _ := syscall.GetStdHandle(syscall.STD_ERROR_HANDLE)  // nolint:errcheck

	// attempt to attach to parent console
	err := procAttachConsole.Find()
	if err != nil {
		log.Err(err).
			Msg("unable to find AttachConsole proc")
		return false, err
	}
	r1, r2, errno := procAttachConsole.Call(windowsAttachParentProcess)
	if r1 == 0 {
		// possible errors:
		// ERROR_ACCESS_DENIED: already attached to console
		// ERROR_INVALID_HANDLE: process does not have console
		// ERROR_INVALID_PARAMETER: process does not exist
		log.Err(errno).
			Uint("r1", uint(r1)).
			Uint("r2", uint(r2)).
			Msg("error attaching to parent console")
		return false, errno
	}

	// get std handles after we attached to console
	var invalid syscall.Handle
	con := invalid

	if stdin == invalid {
		stdin, _ = syscall.GetStdHandle(syscall.STD_INPUT_HANDLE) // nolint:errcheck
	}
	if stdout == invalid {
		stdout, _ = syscall.GetStdHandle(syscall.STD_OUTPUT_HANDLE) // nolint:errcheck
		con = stdout
	}
	if stderr == invalid {
		stderr, _ = syscall.GetStdHandle(syscall.STD_ERROR_HANDLE) // nolint:errcheck
		con = stderr
	}

	// correct output mode
	err = nil
	if con != invalid {
		err = correctOutputMode(con)
	}
	if err != nil {
		log.Err(err).
			Msg("error correcting output mode")
		return false, err
	}

	// fix std handles to correct values (ie. redirects)
	if stdin != invalid {
		os.Stdin = os.NewFile(uintptr(stdin), "stdin")
		log.Debug().
			Msg("fixed os.Stdin after attaching to parent console")
	}
	if stdout != invalid {
		os.Stdout = os.NewFile(uintptr(stdout), "stdout")
		log.Debug().
			Msg("fixed os.Stdout after attaching to parent console")
	}
	if stderr != invalid {
		os.Stderr = os.NewFile(uintptr(stderr), "stderr")
		log.Debug().
			Msg("fixed os.Stderr after attaching to parent console")
	}
	return true, nil
}

func correctOutputMode(con syscall.Handle) error {
	log := logger.GetFuncLogger(packageLogger, "correctOutputMode")
	// Make sure the console is configured to convert
	// \n to \r\n, like Go programs expect.
	h := windows.Handle(con)
	var st uint32
	err := windows.GetConsoleMode(h, &st)
	if err != nil {
		log.Err(err).
			Msg("failed to get console mode")
		return err
	}
	err = windows.SetConsoleMode(h, st&^windows.DISABLE_NEWLINE_AUTO_RETURN)
	if err != nil {
		log.Err(err).
			Msg("failed to set console mode")
		return err
	}
	return nil
}
