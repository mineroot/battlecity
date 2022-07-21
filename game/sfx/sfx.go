package sfx

import (
	"embed"
	"fmt"
	"github.com/faiface/beep"
	"github.com/faiface/beep/speaker"
	"github.com/faiface/beep/wav"
	"os"
	"path/filepath"
	"time"
)

var (
	collection            map[string]*beep.Buffer
	sr                    beep.SampleRate
	tankIdleStream        *beep.Ctrl
	tankMovingStream      *beep.Ctrl
	startUpStream         beep.StreamSeeker
	shootStream           beep.StreamSeeker
	bonusAppearedStream   beep.StreamSeeker
	bonusTakenLifeStream  beep.StreamSeeker
	bonusTakenOtherStream beep.StreamSeeker
	pauseStream           *streamSeekerCtrl
	startUpDone           chan struct{}
)

func Init(files embed.FS) error {
	dirEntries, _ := files.ReadDir("assets/sfx")
	collection = make(map[string]*beep.Buffer)
	for _, fileInfo := range dirEntries {
		if fileInfo.IsDir() || filepath.Ext(fileInfo.Name()) != ".wav" {
			continue
		}
		wavFile, err := os.Open(fmt.Sprintf("assets/sfx/%s", fileInfo.Name()))
		if err != nil {
			return err
		}

		streamer, format, err := wav.Decode(wavFile)
		if err != nil {
			return err
		}
		buffer := beep.NewBuffer(format)
		buffer.Append(streamer)
		_ = streamer.Close()
		collection[fileInfo.Name()] = buffer
	}

	sr = beep.SampleRate(44100)
	_ = speaker.Init(sr, sr.N(time.Second/10))

	startUpStream = stream("StartUp.wav")
	tankIdleStream = &beep.Ctrl{
		Streamer: beep.Loop(-1, stream("TankIdle.wav")),
		Paused:   true,
	}
	tankMovingStream = &beep.Ctrl{
		Streamer: beep.Loop(-1, stream("TankMoving.wav")),
		Paused:   true,
	}
	pauseStream = &streamSeekerCtrl{
		streamer: stream("Pause.wav"),
		Paused:   true,
	}
	shootStream = stream("Shoot.wav")
	bonusAppearedStream = stream("BonusAppeared.wav")
	bonusTakenLifeStream = stream("BonusTakenLife.wav")
	bonusTakenOtherStream = stream("BonusTakenOther.wav")

	return nil
}

func ResetForNewStage() {
	speaker.Clear() // clear all Streamers
	startUpDone = make(chan struct{})
	_ = startUpStream.Seek(0) // rewind startup stream to start
	startUpStreamRewound := beep.Seq(
		beep.Take(sr.N(time.Millisecond*4500), startUpStream),
		beep.Callback(func() {
			close(startUpDone) // startup is done
		}),
	)
	speaker.Play(startUpStreamRewound)
	speaker.Play(pauseStream)

	go func() {
		<-startUpDone // wait for startup is done
		speaker.Play(tankIdleStream, tankMovingStream)
	}()
}

func PlayTankMoving() {
	if !tankMovingStream.Paused {
		return
	}
	speaker.Lock()
	defer speaker.Unlock()
	tankIdleStream.Paused = true
	tankMovingStream.Paused = false
}

func PlayTankIdle() {
	if !tankIdleStream.Paused {
		return
	}
	speaker.Lock()
	defer speaker.Unlock()
	tankMovingStream.Paused = true
	tankIdleStream.Paused = false
}

func PlayBotDestroyed() {
	speaker.Play(beep.Take(sr.N(time.Millisecond*500), stream("BotDestruction.wav")))
}

func PlayPlayerDestroyed() {
	speaker.Play(beep.Take(sr.N(time.Millisecond*500), stream("PlayerDestruction.wav")))
}

func PlayShoot() {
	speaker.Lock()
	_ = shootStream.Seek(0)
	speaker.Unlock()
	speaker.Play(beep.Take(sr.N(time.Millisecond*100), shootStream))
}
func PlayBonusAppeared() {
	speaker.Lock()
	_ = bonusAppearedStream.Seek(0)
	speaker.Unlock()
	speaker.Play(beep.Take(sr.N(time.Millisecond*500), bonusAppearedStream))
}

func PlayBonusTakenLife() {
	speaker.Lock()
	_ = bonusTakenLifeStream.Seek(0)
	speaker.Unlock()
	speaker.Play(beep.Take(sr.N(time.Second), bonusTakenLifeStream))
}

func PlayBonusTakenOther() {
	speaker.Lock()
	_ = bonusTakenOtherStream.Seek(0)
	speaker.Unlock()
	speaker.Play(beep.Take(sr.N(time.Millisecond*700), bonusTakenOtherStream))
}

func PlayPause() {
	speaker.Lock()
	defer speaker.Unlock()
	_ = pauseStream.Seek(0)
	pauseStream.Paused = false
	tankMovingStream.Paused = true
	tankIdleStream.Paused = true
}

func StopPause() {
	speaker.Lock()
	defer speaker.Unlock()
	pauseStream.Paused = true
}

func stream(name string) beep.StreamSeeker {
	buffer := collection[name]
	return buffer.Streamer(0, buffer.Len())
}
