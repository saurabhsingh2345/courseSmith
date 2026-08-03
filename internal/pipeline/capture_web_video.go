package pipeline

// Web video: the shots where time is the subject.
//
// Stills carry most of a no-code course, and capture_web.go argues why. They
// cannot carry the one shot the course exists for — a sentence going in and an
// application coming out. Nobody believes that from two screenshots; the whole
// point is watching it happen, including how long it takes.
//
// == Real timestamps, for once ==
//
// The terminal path has to *model* where its marks fall, because VHS emits no
// timestamps and the whole apparatus in footage.go exists to admit when that
// model cannot be trusted. Here we hold the clock: a screencast frame arrives
// when it arrives, a mark is recorded against the same clock, and the offset is
// measured rather than derived. Web video marks are never approximate, and that
// is not a happy accident — it is the difference between driving the recorder
// and shelling out to one.
//
// == Why frames and ffmpeg rather than a video stream ==
//
// CDP screencast hands back JPEG frames with metadata, not a container. It also
// only sends a frame when something *changed*, which is exactly right for this:
// ninety seconds of a spinner is a handful of frames, so the recording is small
// and the timing information is preserved in the per-frame durations rather
// than in a wall of identical frames. ffmpeg's concat demuxer turns that back
// into a video with honest pacing.

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
)

// screencastQuality trades file size against how readable small UI text is.
// A capture is going to be scaled onto a 1080p stage and then possibly zoomed,
// so this sits higher than a screencast would normally need.
const screencastQuality = 90

// minFrameMs stops a burst of frames in the same millisecond producing a
// zero-duration entry, which ffmpeg's concat demuxer silently drops.
const minFrameMs = 16

// screencastFrame is one captured frame and when it arrived.
type screencastFrame struct {
	data  []byte
	atMs  int
	index int
}

// webRecorder accumulates screencast frames and timed marks for one take.
type webRecorder struct {
	page    *rod.Page
	started time.Time
	marks   []FootageMark
	running bool

	// mu guards frames, which the event subscription appends to from its own
	// goroutine while the take's steps run on this one.
	mu     sync.Mutex
	frames []screencastFrame

	// cancel stops the subscription. rod's EachEvent returns a function that
	// *waits* for events rather than one that cancels — calling it to "stop"
	// blocks until the context dies, which is a three-minute hang followed by
	// an encode against an already-expired context. Cancelling a derived page
	// context is the way to end a background subscription.
	cancel context.CancelFunc
}

// nowMs is the offset into the recording.
func (r *webRecorder) nowMs() int {
	return int(time.Since(r.started).Milliseconds())
}

// start begins the screencast. Frames stream in on a background goroutine that
// rod manages; every one must be acknowledged or Chromium stops sending.
func (r *webRecorder) start() error {
	quality := screencastQuality
	if err := (proto.PageStartScreencast{
		Format:  proto.PageStartScreencastFormatJpeg,
		Quality: &quality,
	}).Call(r.page); err != nil {
		return fmt.Errorf("starting the screencast: %w", err)
	}
	r.started = time.Now()
	r.running = true

	ctx, cancel := context.WithCancel(context.Background())
	r.cancel = cancel
	wait := r.page.Context(ctx).EachEvent(func(e *proto.PageScreencastFrame) {
		r.mu.Lock()
		r.frames = append(r.frames, screencastFrame{
			data:  e.Data,
			atMs:  r.nowMs(),
			index: len(r.frames),
		})
		r.mu.Unlock()
		// Unacknowledged frames stall the stream, and a stalled stream looks
		// exactly like a page that stopped changing.
		_ = (proto.PageScreencastFrameAck{SessionID: e.SessionID}).Call(r.page)
	})
	go wait()
	return nil
}

// finish stops the screencast and returns the frames.
func (r *webRecorder) finish() error {
	if !r.running {
		return nil
	}
	r.running = false
	if err := (proto.PageStopScreencast{}).Call(r.page); err != nil {
		return fmt.Errorf("stopping the screencast: %w", err)
	}
	if r.cancel != nil {
		r.cancel()
	}
	return nil
}

// mark records a named moment against the recording's own clock.
func (r *webRecorder) mark(name string) {
	if !r.running {
		return
	}
	r.marks = append(r.marks, FootageMark{Name: name, AtMs: r.nowMs()})
}

// encode writes the frames to disk and muxes them into an mp4 whose pacing
// matches when they really arrived.
//
// The last frame needs an explicit duration and a repeat of its own filename:
// the concat demuxer takes a file's duration from the *next* entry, so without
// the repeat the final frame is shown for an instant and the recording appears
// to end early — which for a build-an-app clip means losing the finished app.
func (r *webRecorder) encode(ctx context.Context, e *Env, outPath string) (int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if len(r.frames) == 0 {
		return 0, fmt.Errorf("the screencast produced no frames — the page may never have rendered")
	}
	dir, err := os.MkdirTemp("", "coursesmith-screencast-*")
	if err != nil {
		return 0, fmt.Errorf("creating a frame directory: %w", err)
	}
	defer os.RemoveAll(dir)

	var list strings.Builder
	for i, f := range r.frames {
		name := fmt.Sprintf("f-%06d.jpg", i)
		if err := os.WriteFile(filepath.Join(dir, name), f.data, 0o644); err != nil {
			return 0, fmt.Errorf("writing frame %d: %w", i, err)
		}
		durMs := minFrameMs
		if i+1 < len(r.frames) {
			if d := r.frames[i+1].atMs - f.atMs; d > durMs {
				durMs = d
			}
		} else {
			durMs = 1200 // hold the final state so it can be read
		}
		fmt.Fprintf(&list, "file '%s'\nduration %.3f\n", name, float64(durMs)/1000)
	}
	fmt.Fprintf(&list, "file '%s'\n", fmt.Sprintf("f-%06d.jpg", len(r.frames)-1))

	listPath := filepath.Join(dir, "frames.txt")
	if err := os.WriteFile(listPath, []byte(list.String()), 0o644); err != nil {
		return 0, err
	}
	if err := os.MkdirAll(filepath.Dir(outPath), 0o755); err != nil {
		return 0, err
	}
	// yuv420p and even dimensions, because a screencast is whatever size the
	// viewport was and an odd width makes h264 refuse.
	if err := e.runFFmpeg(ctx,
		"-y", "-f", "concat", "-safe", "0", "-i", listPath,
		"-vf", "scale=trunc(iw/2)*2:trunc(ih/2)*2",
		"-c:v", "libx264", "-preset", "medium", "-crf", "20",
		"-pix_fmt", "yuv420p", "-r", "30",
		outPath,
	); err != nil {
		return 0, fmt.Errorf("encoding the screencast: %w", err)
	}
	return mediaDurationMs(outPath)
}

// frameCount is how many frames were captured, safely.
func (r *webRecorder) frameCount() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return len(r.frames)
}
