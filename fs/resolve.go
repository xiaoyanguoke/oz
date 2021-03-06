package fs

import (
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path"
	"path/filepath"
	"strings"
)

func ResolvePathNoGlob(p string, u *user.User) (string, error) {
	return resolveVars(p, u)
}

func resolvePath(p string, u *user.User) ([]string, error) {
	p, err := resolveVars(p, u)
	if err != nil {
		return nil, err
	}
	return resolveGlob(p)
}

func resolveVars(p string, u *user.User) (string, error) {
	const pathVar = "${PATH}/"
	const homeVar = "${HOME}"
	const uidVar = "${UID}"
	const userVar = "${USER}"

	switch {
	case strings.HasPrefix(p, pathVar):
		emptyPath := false
		if os.Getenv("PATH") == "" {
			emptyPath = true
			os.Setenv("PATH", "/bin:/usr/bin:/sbin:/usr/sbin")
		}
		resolved, err := exec.LookPath(p[len(pathVar):])
		if emptyPath {
			os.Setenv("PATH", "") // Do not use Unsetenv, incompatible with golang 1.3
		}
		if err != nil {
			return "", fmt.Errorf("failed to resolve %s", p)
		}
		return resolved, nil

	case strings.HasPrefix(p, homeVar):
		if u == nil {
			return p, nil
		}
		return path.Join(u.HomeDir, p[len(homeVar):]), nil

	case strings.Contains(p, uidVar):
		if u == nil {
			return p, nil
		}
		return strings.Replace(p, uidVar, u.Uid, -1), nil

	case strings.Contains(p, userVar):
		if u == nil {
			return p, nil
		}
		return strings.Replace(p, userVar, u.Username, -1), nil
	}
	return p, nil
}

func isGlobbed(p string) bool {
	return strings.Contains(p, "*")
}

func resolveGlob(p string) ([]string, error) {
	if !isGlobbed(p) {
		return []string{p}, nil
	}
	list, err := filepath.Glob(p)
	if err != nil {
		return nil, fmt.Errorf("failed to glob resolve %s: %v", p, err)
	}
	return list, nil
}
