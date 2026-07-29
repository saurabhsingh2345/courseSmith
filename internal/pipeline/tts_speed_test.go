package pipeline

import "testing"

func TestComputeSpeedFix(t *testing.T) {
	tests := []struct {
		name     string
		measured float64
		target   int
		oldFix   float64
		want     float64 // 0 = nil expected
	}{
		{"in band", 150, 145, 1, 0},
		{"no target", 300, 0, 1, 0},
		{"no measurement", 0, 145, 1, 0},
		{"too fast", 180, 145, 1, 145.0 / 180},
		{"too slow", 110, 145, 1, 145.0 / 110},
		{"clamped fast", 400, 145, 1, minSpeedFix},
		{"clamped slow", 50, 145, 1, maxSpeedFix},
		{"stable at clamp", 260, 145, minSpeedFix, 0}, // can't go lower — no churn
		{"composes with old fix", 180, 145, 1.1, 1.1 * 145.0 / 180},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := computeSpeedFix(tt.measured, tt.target, tt.oldFix)
			if tt.want == 0 {
				if got != nil {
					t.Fatalf("want nil fix, got %+v", got)
				}
				return
			}
			if got == nil {
				t.Fatalf("want speed %.3f, got nil", tt.want)
			}
			if diff := got.Speed - tt.want; diff > 0.001 || diff < -0.001 {
				t.Errorf("speed = %.3f, want %.3f", got.Speed, tt.want)
			}
		})
	}
}

func TestEffectivePaceWPM(t *testing.T) {
	tests := []struct {
		name  string
		pace  int
		speed float64
		want  int
	}{
		{"unset speed means natural rate", 175, 0, 175},
		{"1.0 is the natural rate", 175, 1, 175},
		{"0.9 moves the target down with the voice", 175, 0.9, 158},
		{"faster moves it up", 150, 1.2, 180},
		{"no pace target stays no target", 0, 0.9, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := effectivePaceWPM(tt.pace, tt.speed); got != tt.want {
				t.Errorf("effectivePaceWPM(%d, %.2f) = %d, want %d", tt.pace, tt.speed, got, tt.want)
			}
		})
	}
}
