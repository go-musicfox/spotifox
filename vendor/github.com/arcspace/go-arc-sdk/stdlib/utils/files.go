package utils

import (
	"fmt"
	"io"
	"os"
	"path"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/arcspace/go-arc-sdk/stdlib/bufs"
	"github.com/arcspace/go-arc-sdk/stdlib/errors"
)

var (
	// DefaultDirPerms expresses the default mode of dir creation.
	DefaultDirPerms = os.FileMode(0755)
)

// EnsureDirAndMaxPerms ensures that the given path exists, that it's a directory,
// and that it has permissions that are no more permissive than the given ones.
//
// - If the path does not exist, it is created
// - If the path exists, but is not a directory, an error is returned
// - If the path exists, and is a directory, but has the wrong perms, it is chmod'ed
func EnsureDirAndMaxPerms(path string, perms os.FileMode) error {
	stat, err := os.Stat(path)
	if err != nil && !os.IsNotExist(err) {
		// Regular error
		return err
	} else if os.IsNotExist(err) {
		// Dir doesn't exist, create it with desired perms
		return os.MkdirAll(path, perms)
	} else if !stat.IsDir() {
		// Path exists, but it's a file, so don't clobber
		return errors.Errorf("%v already exists and is not a directory", path)
	} else if stat.Mode() != perms {
		// Dir exists, but wrong perms, so chmod
		return os.Chmod(path, (stat.Mode() & perms))
	}
	return nil
}

// WriteFileWithMaxPerms writes `data` to `path` and ensures that
// the file has permissions that are no more permissive than the given ones.
func WriteFileWithMaxPerms(path string, data []byte, perms os.FileMode) error {
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, perms)
	if err != nil {
		return err
	}
	defer f.Close()
	err = EnsureFileMaxPerms(f, perms)
	if err != nil {
		return err
	}
	_, err = f.Write(data)
	return err
}

// CopyFileWithMaxPerms copies the file at `srcPath` to `dstPath`
// and ensures that it has permissions that are no more permissive than the given ones.
func CopyFileWithMaxPerms(srcPath, dstPath string, perms os.FileMode) error {
	src, err := os.Open(srcPath)
	if err != nil {
		return errors.Wrap(err, "could not open source file")
	}
	defer src.Close()

	dst, err := os.OpenFile(dstPath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, perms)
	if err != nil {
		return errors.Wrap(err, "could not open destination file")
	}
	defer dst.Close()

	err = EnsureFileMaxPerms(dst, perms)
	if err != nil {
		return errors.Wrap(err, "could not set file permissions")
	}

	_, err = io.Copy(dst, src)
	return errors.Wrap(err, "could not copy file contents")
}

// EnsureFileMaxPerms ensures that the given file has permissions
// that are no more permissive than the given ones.
func EnsureFileMaxPerms(file *os.File, perms os.FileMode) error {
	stat, err := file.Stat()
	if err != nil {
		return err
	}
	if stat.Mode() == perms {
		return nil
	}
	return file.Chmod(stat.Mode() & perms)
}

// EnsureFilepathMaxPerms ensures that the file at the given filepath
// has permissions that are no more permissive than the given ones.
func EnsureFilepathMaxPerms(filepath string, perms os.FileMode) error {
	dst, err := os.OpenFile(filepath, os.O_RDWR, perms)
	if err != nil {
		return err
	}
	defer dst.Close()

	return EnsureFileMaxPerms(dst, perms)
}

type FileSize uint64

var fsregex = regexp.MustCompile(`(\d+\.?\d*)(tb|gb|mb|kb|b)?`)

const (
	KB FileSize = 1024
	MB FileSize = 1000000
	GB FileSize = 1000 * MB
	TB FileSize = 1000 * GB
)

func ParseFileSize(s string) (FileSize, error) {
	var fs FileSize
	err := fs.UnmarshalText([]byte(s))
	return fs, err
}

func (s FileSize) MarshalText() ([]byte, error) {
	if s > TB {
		return []byte(fmt.Sprintf("%.2ftb", float64(s)/float64(TB))), nil
	} else if s > GB {
		return []byte(fmt.Sprintf("%.2fgb", float64(s)/float64(GB))), nil
	} else if s > MB {
		return []byte(fmt.Sprintf("%.2fmb", float64(s)/float64(MB))), nil
	} else if s > KB {
		return []byte(fmt.Sprintf("%.2fkb", float64(s)/float64(KB))), nil
	}
	return []byte(fmt.Sprintf("%db", s)), nil
}

func (s *FileSize) UnmarshalText(bs []byte) error {
	matches := fsregex.FindAllStringSubmatch(strings.ToLower(string(bs)), -1)
	if len(matches) != 1 {
		return errors.Errorf(`bad filesize: "%v"`, string(bs))
	} else if len(matches[0]) != 3 {
		return errors.Errorf(`bad filesize: "%v"`, string(bs))
	}
	var (
		num  = matches[0][1]
		unit = matches[0][2]
	)
	bytes, err := strconv.ParseFloat(num, 64)
	if err != nil {
		return err
	}

	switch unit {
	case "", "b":
	case "kb":
		bytes *= float64(KB)
	case "mb":
		bytes *= float64(MB)
	case "gb":
		bytes *= float64(GB)
	case "tb":
		bytes *= float64(TB)
	default:
		return errors.Errorf(`bad filesize unit: "%v"`, unit)
	}
	*s = FileSize(bytes)
	return nil
}

func (s FileSize) String() string {
	str, _ := s.MarshalText()
	return string(str)
}

// ExpandAndCheckPath parses/expands the given path and then verifies it's existence or non-existence,
// depending on autoCreate and returning the the expanded path.
//
// If autoCreate == true, an error is returned if the dir didn't exist and failed to be created.
//
// If autoCreate == false, an error is returned if the dir doesn't exist.
func ExpandAndCheckPath(
	pathname string,
	autoCreate bool,
) (string, error) {

	var err error
	if err != nil {
		err = errors.Errorf("error expanding '%s'", pathname)
	} else {
		_, err = os.Stat(pathname)
		if err != nil && os.IsNotExist(err) {
			if autoCreate {
				err = os.MkdirAll(pathname, DefaultDirPerms)
			} else {
				err = errors.Errorf("path '%s' does not exist", pathname)
			}
		}
	}

	if err != nil {
		return "", err
	}

	return pathname, nil
}

// CreateNewDir creates the specified dir (and returns an error if the dir already exists)
//
// If dirPath is absolute then basePath is ignored.
// Returns the effective pathname.
func CreateNewDir(basePath, dirPath string) (string, error) {
	var pathname string

	if path.IsAbs(dirPath) {
		pathname = dirPath
	} else {
		pathname = path.Join(basePath, dirPath)
	}

	if _, err := os.Stat(pathname); !os.IsNotExist(err) {
		return "", errors.Errorf("for safety, the path '%s' must not already exist", pathname)
	}

	if err := os.MkdirAll(pathname, DefaultDirPerms); err != nil {
		return "", err
	}

	return pathname, nil
}

// GetExePath returns the pathname of the dir containing the host exe
func GetExePath() (string, error) {
	hostExe, err := os.Executable()
	if err != nil {
		return "", err
	}
	hostDir := path.Dir(hostExe)
	return hostDir, nil
}

var remapCharset = map[rune]rune{
	' ':  '-',
	'.':  '-',
	'?':  '-',
	'\\': '+',
	'/':  '+',
	'&':  '+',
}

// MakeFSFriendly makes a given string safe to use for a file system.
// If suffix is given, the hex encoding of those bytes are appended after a space.
func MakeFSFriendly(name string, suffix []byte) string {

	var b strings.Builder
	for _, r := range name {
		if replace, ok := remapCharset[r]; ok {
			if replace != 0 {
				b.WriteRune(replace)
			}
		} else {
			b.WriteRune(r)
		}
	}

	if len(suffix) > 0 {
		b.WriteString(" ")
		b.WriteString(bufs.Base32Encoding.EncodeToString(suffix))
	}

	friendlyName := b.String()
	return friendlyName
}

// CreateTemp creates a temporary file with the given file mode flags using the given name pattern.
// The name pattern should contain a '*' to where a random alphanumeric should go.
func CreateTemp(dir, pattern string, fileFlags int) (ofile *os.File, err error) {
	if dir == "" {
		dir = os.TempDir()
	}

	var prefix, suffix string
	if pos := strings.LastIndexByte(pattern, '*'); pos != -1 {
		prefix, suffix = pattern[:pos], pattern[pos+1:]
	} else {
		prefix = pattern
	}
	prefix = path.Join(dir, prefix)

	for {
		msec := time.Now().UnixMicro()
		name := prefix + strconv.FormatInt(int64(msec), 32) + suffix
		ofile, err = os.OpenFile(name, os.O_CREATE|fileFlags, 0600)
		if os.IsExist(err) {
			continue
		}
		break
	}
	return
}
