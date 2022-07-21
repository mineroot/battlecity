package sfx

import "github.com/faiface/beep"

type streamSeekerCtrl struct {
	streamer beep.StreamSeeker
	Paused   bool
}

func (s *streamSeekerCtrl) Len() int {
	return s.streamer.Len()
}

func (s *streamSeekerCtrl) Position() int {
	return s.streamer.Position()
}

func (s *streamSeekerCtrl) Seek(p int) error {
	return s.streamer.Seek(p)
}

// Stream streams the wrapped streamer, if not nil. If the streamer is nil, Ctrl acts as drained.
// When paused, Ctrl streams silence.
// When streamer drains it rewinds and pauses
func (s *streamSeekerCtrl) Stream(samples [][2]float64) (n int, ok bool) {
	if s.streamer == nil {
		return 0, false
	}
	if s.Paused {
		for i := range samples {
			samples[i] = [2]float64{}
		}
		return len(samples), true
	}
	n, ok = s.streamer.Stream(samples)
	if !ok {
		s.Paused = true
		_ = s.Seek(0)
		n = len(samples)
		ok = true
	}
	return
}

// Err returns the error of the wrapped Streamer, if not nil.
func (s *streamSeekerCtrl) Err() error {
	if s.streamer == nil {
		return nil
	}
	return s.streamer.Err()
}
