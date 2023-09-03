package utils

import (
	"embed"
	"encoding/binary"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path"
	"strconv"
	"time"

	"github.com/go-musicfox/spotifox/internal/configs"
	"github.com/go-musicfox/spotifox/internal/constants"
	"github.com/zmb3/spotify/v2"

	"github.com/buger/jsonparser"
	"golang.org/x/mod/semver"
)

//go:embed embed
var embedDir embed.FS

func GetLocalDataDir() string {
	if root := os.Getenv("SPOTIFOX_ROOT"); root != "" {
		return root
	}
	configDir, err := os.UserConfigDir()
	if nil != err {
		panic("cannot find local storage dir:" + err.Error())
	}
	return path.Join(configDir, constants.AppLocalDataDir)
}

// IDToBin convert autoincrement ID to []byte
func IDToBin(ID uint64) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, ID)
	return b
}

// BinToID convert []byte to autoincrement ID
func BinToID(bin []byte) uint64 {
	ID := binary.BigEndian.Uint64(bin)

	return ID
}

func LoadIniConfig() {
	projectDir := GetLocalDataDir()
	configFile := path.Join(projectDir, constants.AppIniFile)
	if !FileOrDirExists(configFile) {
		_ = CopyFileFromEmbed("embed/spotifox.ini", configFile)
	}
	configs.ConfigRegistry = configs.NewRegistryFromIniFile(configFile)
}

func CheckUpdate() (bool, string) {
	response, err := http.Get(constants.AppCheckUpdateUrl)
	if err != nil {
		return false, ""
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(response.Body)

	jsonBytes, err := io.ReadAll(response.Body)
	if err != nil {
		return false, ""
	}

	tag, err := jsonparser.GetString(jsonBytes, "tag_name")
	if err != nil {
		return false, ""
	}

	return semver.Compare(tag, constants.AppVersion) > 0, tag
}

func CopyFileFromEmbed(src, dst string) error {
	var (
		err   error
		srcfd fs.File
		dstfd *os.File
	)

	if srcfd, err = embedDir.Open(src); err != nil {
		return err
	}
	defer srcfd.Close()

	if dstfd, err = os.OpenFile(dst, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0766); err != nil {
		return err
	}
	defer dstfd.Close()

	if _, err = io.Copy(dstfd, srcfd); err != nil {
		return err
	}
	return nil
}

func CopyDirFromEmbed(src, dst string) error {
	var (
		err error
		fds []fs.DirEntry
	)

	if err = os.MkdirAll(dst, 0766); err != nil {
		return err
	}
	if fds, err = embedDir.ReadDir(src); err != nil {
		return err
	}
	for _, fd := range fds {
		srcfp := path.Join(src, fd.Name())
		dstfp := path.Join(dst, fd.Name())

		if fd.IsDir() {
			if err = CopyDirFromEmbed(srcfp, dstfp); err != nil {
				return err
			}
		} else {
			if err = CopyFileFromEmbed(srcfp, dstfp); err != nil {
				return err
			}
		}
	}
	return nil
}

func WebUrlOfPlaylist(playlistId int64) string {
	return "https://music.163.com/#/my/m/music/playlist?id=" + strconv.FormatInt(playlistId, 10)
}

func WebUrlOfSong(songId spotify.ID) string {
	return "https://open.spotify.com/track/" + string(songId)
}

func WebUrlOfArtist(artistId int64) string {
	return "https://music.163.com/#/artist?id=" + strconv.FormatInt(artistId, 10)
}

func WebUrlOfAlbum(artistId int64) string {
	return "https://music.163.com/#/album?id=" + strconv.FormatInt(artistId, 10)
}

func FileUrl(filepath string) string {
	return "file://" + filepath
}

func IsSameDate(t1, t2 time.Time) bool {
	y1, m1, d1 := t1.Date()
	y2, m2, d2 := t2.Date()
	return y1 == y2 && m1 == m2 && d1 == d2
}

func FileOrDirExists(filename string) bool {
	_, err := os.Stat(filename)
	return !os.IsNotExist(err)
}
