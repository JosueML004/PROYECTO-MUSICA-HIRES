package main

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"time"

	"github.com/dhowden/tag"
	"github.com/faiface/beep"
	"github.com/faiface/beep/effects"
	"github.com/faiface/beep/flac"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
)

var (
	ctrl            *beep.Ctrl
	currentStreamer beep.StreamSeekCloser
	currentFormat   beep.Format
	volume          *effects.Volume
)

type MusicFile struct {
	Path     string
	Artist   string
	Album    string
	Title    string
	Duration time.Duration
}

// ScanMusicDirectory scans the given directory recursively for music files and extracts metadata
func ScanMusicDirectory(root string) ([]MusicFile, error) {
	var musicFiles []MusicFile

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && (filepath.Ext(path) == ".flac" || filepath.Ext(path) == ".mp3") {
			file, err := os.Open(path)
			if err != nil {
				fmt.Printf("Error opening file %s: %v\n", path, err)
				return nil // Don't stop walking on individual file errors
			}
			defer file.Close()

			var title, artist, album string

			m, err := tag.ReadFrom(file)
			if err != nil {
				// If reading tags fails, use the filename as the title
				fmt.Printf("Error reading tags from %s: %v. Using filename as title.\n", path, err)
				title = filepath.Base(path)
				artist = "Unknown Artist"
				album = "Unknown Album"
			} else {
				// If successful, use the read metadata
				title = m.Title()
				artist = m.Artist()
				album = m.Album()
			}

			// Reset file cursor to the beginning before decoding for duration
			_, err = file.Seek(0, 0)
			if err != nil {
				fmt.Printf("Error seeking file %s: %v\n", path, err)
				return nil
			}

			var duration time.Duration
			if filepath.Ext(path) == ".mp3" {
				streamer, format, err := mp3.Decode(file)
				if err == nil {
					duration = format.SampleRate.D(streamer.Len())
					streamer.Close()
				}
			} else if filepath.Ext(path) == ".flac" {
				streamer, format, err := flac.Decode(file)
				if err == nil {
					duration = format.SampleRate.D(streamer.Len())
					streamer.Close()
				}
			}

			musicFiles = append(musicFiles, MusicFile{
				Artist:   artist,
				Album:    album,
				Title:    title,
				Duration: duration,
				Path:     path,
			})
		}
		return nil
	})

	if err != nil {
		return nil, err
	}
	return musicFiles, nil
}

func PlayMusicFile(filePath string) error {
	StopMusic() // Stop any currently playing music

	f, err := os.Open(filePath)
	if err != nil {
		return err
	}

	var s beep.StreamSeekCloser
	var format beep.Format

	switch filepath.Ext(filePath) {
	case ".mp3":
		s, format, err = mp3.Decode(f)
	case ".flac":
		s, format, err = flac.Decode(f)
	default:
		f.Close()
		return fmt.Errorf("unsupported file format: %s", filepath.Ext(filePath))
	}

	if err != nil {
		f.Close()
		return err
	}

	speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10))

	currentStreamer = s
	currentFormat = format
	// Use effects.Volume to control the volume
	volume = &effects.Volume{Streamer: s, Base: 2, Volume: 0, Silent: false}
	ctrl = &beep.Ctrl{Streamer: volume, Paused: false}

	speaker.Play(ctrl)

	fmt.Printf("Reproduciendo: %s\n", filepath.Base(filePath))
	return nil
}

func PauseMusic() {
	if ctrl != nil {
		speaker.Lock()
		ctrl.Paused = true
		speaker.Unlock()
		fmt.Println("Música en pausa.")
	}
}

func ResumeMusic() {
	if ctrl != nil {
		speaker.Lock()
		ctrl.Paused = false
		speaker.Unlock()
		fmt.Println("Música reanudada.")
	}
}

func StopMusic() {
	if currentStreamer != nil {
		speaker.Lock()
		currentStreamer.Close()
		currentStreamer = nil
		ctrl = nil
		speaker.Unlock()
		fmt.Println("Reproducción detenida.")
	}
}

func SeekMusic(duration time.Duration) {
	if currentStreamer != nil {
		speaker.Lock()
		newPos := currentStreamer.Position() + currentFormat.SampleRate.N(duration)
		if newPos < 0 {
			newPos = 0
		}
		if newPos > currentStreamer.Len() {
			newPos = currentStreamer.Len()
		}
		currentStreamer.Seek(newPos)
		speaker.Unlock()
		fmt.Printf("Saltando %v segundos.\n", duration.Seconds())
	}
}

func TogglePlayPause() {
	if ctrl != nil {
		speaker.Lock()
		ctrl.Paused = !ctrl.Paused
		speaker.Unlock()
		if ctrl.Paused {
			fmt.Println("Música pausada.")
		} else {
			fmt.Println("Música reanudada.")
		}
	}
}

// SetVolume adjusts the volume in a natural, logarithmic way.
func SetVolume(vol float64) {
	if volume != nil {
		speaker.Lock()
		// A volume of 0.0 should be silent.
		// A volume of 1.0 should be the original volume.
		if vol <= 0 {
			volume.Silent = true
		} else {
			volume.Silent = false
			// Map linear volume [0, 1] to logarithmic volume.
			// `Base` is 2, so `Volume` is Log2 of the gain.
			// A gain of 1 (vol=1) means Volume=0.
			// A gain of 0.5 (vol=0.5) means Volume=-1.
			volume.Volume = math.Log2(vol)
		}
		speaker.Unlock()
		fmt.Printf("Volumen ajustado a: %.2f\n", vol)
	}
}
