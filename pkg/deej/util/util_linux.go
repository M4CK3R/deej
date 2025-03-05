package util

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/godbus/dbus/v5"
	"github.com/jezek/xgb"
	"github.com/jezek/xgb/xproto"
)

func isWayland() bool {
	// Check if WAYLAND_DISPLAY is set
	if os.Getenv("WAYLAND_DISPLAY") != "" {
		return true
	}

	// Check if XDG_SESSION_TYPE is set to "wayland"
	if os.Getenv("XDG_SESSION_TYPE") == "wayland" {
		return true
	}

	return false
}

func getCurrentWindowProcessNames() ([]string, error) {
	if isWayland() {
		return getCurrentWaylandWindowProcessNames()
	}

	return getCurrentXWindowProcessNames()
}

func getCurrentWaylandWindowProcessNames() ([]string, error) {
	conn, err := dbus.SessionBus()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to session bus: %w", err)
	}
	defer conn.Close()
	obj := conn.Object("org.gnome.Shell", "/org/gnome/Shell")
	response := obj.Call("org.gnome.Shell.Eval", 0, `global             
      .get_window_actors()
      .map(a=>a.meta_window)
	  .filter(a=>a.has_focus())                                   
      .map(w=>w.get_pid())`)

	if response.Err != nil {
		return nil, fmt.Errorf("failed to get PID: %w", response.Err)
	}

	pidString := response.Body[1].(string)
	// Remove brackets
	pidString = strings.TrimLeft(pidString, "[")
	pidString = strings.TrimRight(pidString, "]")

	return getProcessNamesFromPid(pidString)
}
func getCurrentXWindowProcessNames() ([]string, error) {

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

	return getProcessNamesFromPid(fmt.Sprintf("%d", pid))
}

func getProcessNamesFromPid(pid string) ([]string, error) {
	// Read process name from /proc/{pid}/exe
	path, err := os.Readlink(filepath.Join("/proc", pid, "exe"))
	if err != nil {
		return nil, fmt.Errorf("failed to read process path: %w", err)
	}

	processName := filepath.Base(path)

	return []string{string(processName), processName}, nil
}
