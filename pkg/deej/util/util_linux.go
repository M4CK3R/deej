package util

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/jezek/xgb"
	"github.com/jezek/xgb/xproto"
)

func getCurrentWindowProcessNames() ([]string, error) {
	X, err := xgb.NewConn()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to X server: %w", err)
	}
	defer X.Close()

	root := xproto.Setup(X).DefaultScreen(X).Root
	activeAtom, err := xproto.InternAtom(X, true, uint16(len("_NET_ACTIVE_WINDOW")), "_NET_ACTIVE_WINDOW").Reply()
	if err != nil {
		return nil, fmt.Errorf("failed to get active window atom: %w", err)
	}

	reply, err := xproto.GetProperty(X, false, root, activeAtom.Atom,
		xproto.AtomWindow, 0, 1).Reply()
	if err != nil {
		return nil, fmt.Errorf("failed to get active window: %w", err)
	}

	if reply.ValueLen == 0 {
		return []string{}, nil
	}

	windowID := xproto.Window(xgb.Get32(reply.Value))
	pidAtom, err := xproto.InternAtom(X, true, uint16(len("_NET_WM_PID")), "_NET_WM_PID").Reply()
	if err != nil {
		return nil, fmt.Errorf("failed to get PID atom: %w", err)
	}

	pidReply, err := xproto.GetProperty(X, false, windowID, pidAtom.Atom,
		xproto.AtomCardinal, 0, 1).Reply()
	if err != nil {
		return nil, fmt.Errorf("failed to get window PID: %w", err)
	}

	if pidReply.ValueLen == 0 {
		return []string{}, nil
	}

	pid := int(xgb.Get32(pidReply.Value))

	// Read process name from /proc/{pid}/exe
	path, err := os.Readlink(filepath.Join("/proc", fmt.Sprintf("%d", pid), "exe"))
	if err != nil {
		return nil, fmt.Errorf("failed to read process path: %w", err)
	}

	processName := filepath.Base(path)

	return []string{string(processName), processName}, nil
}
