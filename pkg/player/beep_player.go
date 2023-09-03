package player

import (
	"context"
	"io"
	"net/http"
	"os"
	"path"
	"sync"
	"time"

	"github.com/go-musicfox/spotifox/pkg/configs"
	"github.com/go-musicfox/spotifox/pkg/constants"
	"github.com/go-musicfox/spotifox/utils"
	"github.com/zmb3/spotify/v2"

	"github.com/faiface/beep"
	"github.com/faiface/beep/effects"
	"github.com/faiface/beep/speaker"
)

type beepPlayer struct {
	l sync.Mutex

	curMusic MediaAsset
	timer    *utils.Timer

	cacheReader     *os.File
	cacheWriter     *os.File
	cacheDownloaded bool

	curStreamer beep.StreamSeekCloser
	curFormat   beep.Format

	state      State
	ctrl       *beep.Ctrl
	volume     *effects.Volume
	timeChan   chan time.Duration
	stateChan  chan State
	musicChan  chan MediaAsset
	httpClient *http.Client

	close chan struct{}
}

func NewBeepPlayer() Player {
	p := &beepPlayer{
		state: Stopped,

		timeChan:  make(chan time.Duration),
		stateChan: make(chan State),
		musicChan: make(chan MediaAsset),
		ctrl: &beep.Ctrl{
			Paused: false,
		},
		volume: &effects.Volume{
			Base:   2,
			Silent: false,
		},
		httpClient: &http.Client{},
		close:      make(chan struct{}),
	}

	go utils.PanicRecoverWrapper(false, p.listen)

	return p
}

// listen 开始监听
func (p *beepPlayer) listen() {
	var (
		done       = make(chan struct{})
		reader     io.ReadCloser
		err        error
		ctx        context.Context
		cancel     context.CancelFunc
		prevSongId spotify.ID
		doneHandle = func() {
			select {
			case done <- struct{}{}:
			case <-p.close:
			}
		}
	)

	cacheFile := path.Join(utils.GetLocalDataDir(), "music_cache")
	for {
		select {
		case <-p.close:
			if cancel != nil {
				cancel()
			}
			return
		case <-done:
			p.Stop()
		case p.curMusic = <-p.musicChan:
			p.l.Lock()
			p.pausedNoLock()
			if p.timer != nil {
				p.timer.SetPassed(0)
			}
			// 清理上一轮
			if cancel != nil {
				cancel()
			}
			p.reset()
			if prevSongId != p.curMusic.SongInfo.ID || !utils.FileOrDirExists(cacheFile) {
				ctx, cancel = context.WithCancel(context.Background())

				// FIXME No other optimization methods found
				if p.cacheReader, err = os.OpenFile(cacheFile, os.O_CREATE|os.O_TRUNC|os.O_RDONLY, 0666); err != nil {
					panic(err)
				}
				if p.cacheWriter, err = os.OpenFile(cacheFile, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0666); err != nil {
					panic(err)
				}

				if reader, err = p.curMusic.NewAssetReader(); err != nil {
					utils.Logger().Printf("new asset reader err: %+v", err)
					p.stopNoLock()
					continue
				}

				go func(ctx context.Context, cacheWFile *os.File, read io.ReadCloser) {
					defer func() {
						if utils.Recover(true) {
							p.Stop()
						}
					}()
					_, _ = utils.CopyClose(ctx, cacheWFile, read)
					p.cacheDownloaded = true
					p.l.Lock()
					defer p.l.Unlock()
					if p.curStreamer == nil {
						// nil说明外层解析还没开始或解析失败，这里直接退出
						return
					}
					// 除了MP3格式，其他格式无需重载
					if p.curMusic.SongType() == Mp3 && configs.ConfigRegistry.PlayerBeepMp3Decoder != constants.BeepMiniMp3Decoder {
						// 需再开一次文件，保证其指针变化，否则将概率导致 p.ctrl.Streamer = beep.Seq(……) 直接停止播放
						cacheReader, _ := os.OpenFile(cacheFile, os.O_RDONLY, 0666)
						// 使用新的文件后需手动Seek到上次播放处
						lastStreamer := p.curStreamer
						defer lastStreamer.Close()
						pos := lastStreamer.Position()
						if p.curStreamer, p.curFormat, err = DecodeSong(p.curMusic.SongType(), cacheReader); err != nil {
							p.stopNoLock()
							return
						}
						if pos >= p.curStreamer.Len() {
							pos = p.curStreamer.Len() - 1
						}
						if pos < 0 {
							pos = 1
						}
						_ = p.curStreamer.Seek(pos)
						p.ctrl.Streamer = beep.Seq(beep.StreamerFunc(p.streamer), beep.Callback(doneHandle))
					}
				}(ctx, p.cacheWriter, reader)

				var N = 512
				if err = utils.WaitForNBytes(p.cacheReader, N, time.Millisecond*100, 50); err != nil {
					utils.Logger().Printf("WaitForNBytes err: %+v", err)
					p.stopNoLock()
					continue
				}
			} else {
				// 单曲循环以及歌单只有一首歌时不再请求网络
				if p.cacheReader, err = os.OpenFile(cacheFile, os.O_RDONLY, 0666); err != nil {
					panic(err)
				}
			}

			if p.curStreamer, p.curFormat, err = DecodeSong(p.curMusic.SongType(), p.cacheReader); err != nil {
				p.stopNoLock()
				break
			}

			if err = speaker.Init(p.curFormat.SampleRate, p.curFormat.SampleRate.N(time.Millisecond*200)); err != nil {
				panic(err)
			}

			p.ctrl.Streamer = beep.Seq(beep.StreamerFunc(p.streamer), beep.Callback(doneHandle))
			p.volume.Streamer = p.ctrl
			speaker.Play(p.volume)

			// 计时器
			p.timer = utils.NewTimer(utils.Options{
				Duration:       8760 * time.Hour,
				TickerInternal: 200 * time.Millisecond,
				OnRun:          func(started bool) {},
				OnPaused:       func() {},
				OnDone:         func(stopped bool) {},
				OnTick: func() {
					select {
					case p.timeChan <- p.timer.Passed():
					default:
					}
				},
			})
			p.resumeNoLock()
			prevSongId = p.curMusic.SongInfo.ID
			p.l.Unlock()
		}
	}
}

// Play 播放音乐
func (p *beepPlayer) Play(music MediaAsset) {
	select {
	case p.musicChan <- music:
	default:
	}
}

func (p *beepPlayer) CurMusic() MediaAsset {
	return p.curMusic
}

func (p *beepPlayer) setState(state State) {
	p.state = state
	select {
	case p.stateChan <- state:
	case <-time.After(time.Second * 2):
	}
}

// State 当前状态
func (p *beepPlayer) State() State {
	return p.state
}

// StateChan 状态发生变更
func (p *beepPlayer) StateChan() <-chan State {
	return p.stateChan
}

func (p *beepPlayer) PassedTime() time.Duration {
	if p.timer == nil {
		return 0
	}
	return p.timer.Passed()
}

// TimeChan 获取定时器
func (p *beepPlayer) TimeChan() <-chan time.Duration {
	return p.timeChan
}

func (p *beepPlayer) Seek(duration time.Duration) {
	if duration < 0 {
		return
	}
	// FIXME: 暂时仅对MP3格式提供跳转功能
	// FLAC格式(其他未测)跳转会占用大量CPU资源，比特率越高占用越高
	// 导致Seek方法卡住20-40秒的时间，之后方可随意跳转
	// minimp3未实现Seek
	if p.curStreamer == nil || p.curMusic.SongType() != Mp3 || configs.ConfigRegistry.PlayerBeepMp3Decoder == constants.BeepMiniMp3Decoder {
		return
	}
	if p.state == Playing || p.state == Paused {
		speaker.Lock()
		newPos := p.curFormat.SampleRate.N(duration)

		if newPos < 0 {
			newPos = 0
		}
		if newPos >= p.curStreamer.Len() {
			newPos = p.curStreamer.Len() - 1
		}
		if p.curStreamer != nil {
			err := p.curStreamer.Seek(newPos)
			if err != nil {
				utils.Logger().Printf("seek error: %+v", err)
			}
		}
		if p.timer != nil {
			p.timer.SetPassed(duration)
		}
		speaker.Unlock()
	}
}

// UpVolume 调大音量
func (p *beepPlayer) UpVolume() {
	if p.volume.Volume >= 0 {
		return
	}
	p.l.Lock()
	defer p.l.Unlock()

	p.volume.Silent = false
	p.volume.Volume += 0.25
}

// DownVolume 调小音量
func (p *beepPlayer) DownVolume() {
	if p.volume.Volume <= -5 {
		return
	}

	p.l.Lock()
	defer p.l.Unlock()

	p.volume.Volume -= 0.25
	if p.volume.Volume <= -5 {
		p.volume.Silent = true
	}
}

func (p *beepPlayer) Volume() int {
	return int((p.volume.Volume + 5) * 100 / 5) // 转为0~100存储
}

func (p *beepPlayer) SetVolume(volume int) {
	if volume > 100 {
		volume = 100
	}
	if volume < 0 {
		volume = 0
	}

	p.l.Lock()
	defer p.l.Unlock()
	p.volume.Volume = float64(volume)*5/100 - 5
}

func (p *beepPlayer) pausedNoLock() {
	if p.state != Playing {
		return
	}
	p.ctrl.Paused = true
	p.timer.Pause()
	p.setState(Paused)
}

// Paused 暂停播放
func (p *beepPlayer) Paused() {
	p.l.Lock()
	defer p.l.Unlock()
	p.pausedNoLock()
}

func (p *beepPlayer) resumeNoLock() {
	if p.state == Playing {
		return
	}
	p.ctrl.Paused = false
	go p.timer.Run()
	p.setState(Playing)
}

// Resume 继续播放
func (p *beepPlayer) Resume() {
	p.l.Lock()
	defer p.l.Unlock()
	p.resumeNoLock()
}

func (p *beepPlayer) stopNoLock() {
	if p.state == Stopped {
		return
	}
	p.ctrl.Paused = true
	p.timer.Pause()
	p.setState(Stopped)
}

// Stop 停止
func (p *beepPlayer) Stop() {
	p.l.Lock()
	defer p.l.Unlock()
	p.stopNoLock()
}

// Toggle 切换状态
func (p *beepPlayer) Toggle() {
	switch p.State() {
	case Paused, Stopped:
		p.Resume()
	case Playing:
		p.Paused()
	}
}

// Close 关闭
func (p *beepPlayer) Close() {
	p.l.Lock()
	defer p.l.Unlock()

	if p.timer != nil {
		p.timer.Stop()
	}
	close(p.close)
	speaker.Clear()
}

func (p *beepPlayer) reset() {
	speaker.Clear()
	speaker.Close()
	// 关闭旧计时器
	if p.timer != nil {
		p.timer.Stop()
	}
	if p.cacheReader != nil {
		_ = p.cacheReader.Close()
	}
	if p.cacheWriter != nil {
		_ = p.cacheWriter.Close()
	}
	if p.curStreamer != nil {
		_ = p.curStreamer.Close()
		p.curStreamer = nil
	}
	p.cacheDownloaded = false
}

func (p *beepPlayer) streamer(samples [][2]float64) (n int, ok bool) {
	n, ok = p.curStreamer.Stream(samples)
	err := p.curStreamer.Err()
	if err == nil && (ok || p.cacheDownloaded) {
		return
	}
	p.pausedNoLock()

	var retry = 4
	for !ok && retry > 0 {
		utils.ResetError(p.curStreamer)

		select {
		case <-time.After(time.Second * 5):
			n, ok = p.curStreamer.Stream(samples)
		case <-p.close:
			return
		}
		retry--
	}
	p.resumeNoLock()
	return
}
