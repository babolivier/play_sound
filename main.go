package main

import (
	"errors"
	"flag"
	"io/ioutil"
	"math/rand"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/hajimehoshi/oto"
	"github.com/sirupsen/logrus"
)

var (
	// Command-line arguments/flags definition
	port       = flag.String("p", "8888", "Port to listen on")
	audioDir   = flag.String("d", "", "Path to the directory containing audio files")
	sampleRate = flag.Int("r", 48000, "Sample rate")
	mono       = flag.Bool("m", false, "Use mono instead of stereo")

	// Error returned if no directory has been provided or if it doesn't contain any WAV file.
	errNoAudioDir = errors.New("please provide a path to a directory containing WAV files with -d")

	// Global player to use when playing sound.
	otoPlayer *oto.Player
	// In-memory store of info and contents of the audio files in the provided directory.
	// Must be loaded before setting up the HTTP listener.
	audioFiles []*audioFile
	// Semaphore of sorts to make sure we don't trying to play two files at the same time.
	playing bool
)

// audioFile contains info and the content of a WAV audio file.
type audioFile struct {
	name    string
	content []byte
}

func main() {
	// Initialise the RNG.
	rand.Seed(time.Now().UnixNano())

	// Configure logrus so it logs the full time.
	logrus.SetFormatter(
		&logrus.TextFormatter{
			TimestampFormat: "01/02/2006 - 15:04:05",
			FullTimestamp:   true,
		},
	)

	// Parse command-line flags.
	flag.Parse()

	// Try to load the files in the provided directory.
	if err := loadFiles(); err != nil {
		panic(err)
	}

	// Figure out how many channels to use.
	channelNum := 2
	if *mono {
		channelNum = 1
	}

	// Initialise the oto context with the data provided. We set a size of 1 byte because
	// we're feeding audio content byte by byte in the HTTP handler to allow for more
	// flexibility.
	ctx, err := oto.NewContext(*sampleRate, channelNum, 2, 1)
	if err != nil {
		panic(err)
	}

	// Make sure any resource is freed up when the program exits.
	defer ctx.Close()

	// Create and set the global player.
	otoPlayer = ctx.NewPlayer()

	// Register the HTTP handler and start the HTTP listener.
	http.HandleFunc("/play", handleReq)
	if err := http.ListenAndServe("0.0.0.0:"+*port, http.DefaultServeMux); err != nil {
		panic(err)
	}
}

// loadFiles attempts to load WAV files from the directory provided in the command-line
// arguments.
// Returns errNoAudioDir if no directory has been provided or if the directory doesn't
// contain any WAV file. Also returns an error if an I/O issue happened when trying to
// read the directory or one of its files.
func loadFiles() error {
	// Check if a directory has been provided.
	if *audioDir == "" {
		return errNoAudioDir
	}

	// Gather info about the files in the directory.
	files, err := ioutil.ReadDir(*audioDir)
	if err != nil {
		return err
	}

	// Iterate over that info.
	for _, f := range files {
		fileName := f.Name()
		// Filter out any file that doesn't end with ".wav" (i.e. that isn't a WAV file).
		if strings.HasSuffix(fileName, ".wav") {
			// Read the file.
			content, err := ioutil.ReadFile(filepath.Join(*audioDir, fileName))
			if err != nil {
				return err
			}

			// Append the file's data to the global in-memory store.
			audioFiles = append(audioFiles, &audioFile{fileName, content})
		}
	}

	// Make sure we've managed to find and load at least one file.
	if len(audioFiles) == 0 {
		return errNoAudioDir
	}

	logrus.Infof("Loaded %d files", len(audioFiles))

	return nil
}

// handleReq handles incoming request by having a goroutine attempt to play a random
// sound, and immediately returning with 200 OK.
func handleReq(w http.ResponseWriter, req *http.Request) {
	go play()

	w.WriteHeader(200)
}

// play attempts to play a random audio file from the in-memory store.
func play() {
	// Don't attempt to play if something is already playing.
	if !playing {
		playing = true

		// Make sure we set the semaphore to the right value whether we exit normally or
		// return because of an issue.
		defer func() { playing = false }()

		// Select a random file from the in-memory store.
		file := audioFiles[rand.Intn(len(audioFiles))]

		logrus.Infof("Playing %s", file.name)

		// Write into the oto player byte by byte.
		for _, b := range file.content {
			n, err := otoPlayer.Write([]byte{b})
			if err != nil || n != 1 {
				logrus.WithError(err).Error("Failed to play %s", file.name)
				return
			}
		}
	}
}
