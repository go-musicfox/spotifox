//go:build darwin

package avcore

import (
	"fmt"
	"testing"
	"time"

	"github.com/go-musicfox/spotifox/pkg/macdriver/cocoa"
	"github.com/go-musicfox/spotifox/pkg/macdriver/core"
)

func TestMain(m *testing.M) {
	app := cocoa.NSApp()
	if app.ID == 0 {
		panic("app init error")
	}

	app.SetActivationPolicy(cocoa.NSApplicationActivationPolicyProhibited)
	app.ActivateIgnoringOtherApps(true)

	go func() {
		m.Run()
		app.Terminate(0)
	}()

	app.Run()
}

func TestAVPlayer(t *testing.T) {
	player := AVPlayer_alloc().Init()
	player.SetActionAtItemEnd(2)
	player.SetVolume(0.1)
	if player.ID == 0 {
		panic("init player failed")
	}
	defer player.Release()

	file := core.String("./testdata/a.mp3")
	defer file.Release()
	url := core.NSURL_fileURLWithPath(file)
	defer url.Release()
	item := AVPlayerItem_playerItemWithURL(url)
	defer item.Release()

	player.ReplaceCurrentItemWithPlayerItem(item)
	player.Play()
	<-time.After(time.Second * 2)
	player.Pause()

	curItem := player.CurrentItem()
	if curItem.ID == 0 {
		panic("get player current item failed")
	}
	asset := curItem.Asset()
	fmt.Println(asset.URL().AbsoluteString())

	curTime := player.CurrentTime()
	fmt.Println(curTime)

	player.SeekToTime(curTime)
}
