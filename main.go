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
	port       = flag.String("p", "8888", "Port to listen on")
	audioDir   = flag.String("d", "", "Path to the directory containing audio files")
	sampleRate = flag.Int("r", 48000, "Sample rate")
	mono       = flag.Bool("m", false, "Use mono instead of stereo")

	errNoAudioDir = errors.New("please provide a path to a directory containing WAV files with -d")

	otoPlayer  *oto.Player
	audioFiles []*audioFile
	playing    bool
)

type audioFile struct {
	name    string
	content []byte
}

func main() {
	rand.Seed(time.Now().UnixNano())

	logrus.SetFormatter(
		&logrus.TextFormatter{
			TimestampFormat: "01/02/2006 - 15:04:05",
			FullTimestamp:   true,
		},
	)

	flag.Parse()

	if err := loadFiles(); err != nil {
		panic(err)
	}

	channelNum := 2
	if *mono {
		channelNum = 1
	}

	ctx, err := oto.NewContext(*sampleRate, channelNum, 2, 1)
	if err != nil {
		panic(err)
	}

	defer ctx.Close()

	otoPlayer = ctx.NewPlayer()

	http.HandleFunc("/play", handleReq)

	if err := http.ListenAndServe("0.0.0.0:"+*port, http.DefaultServeMux); err != nil {
		panic(err)
	}
}

func loadFiles() error {
	if *audioDir == "" {
		return errNoAudioDir
	}

	files, err := ioutil.ReadDir(*audioDir)
	if err != nil {
		return err
	}

	for _, f := range files {
		fileName := f.Name()
		if strings.HasSuffix(fileName, ".wav") {
			content, err := ioutil.ReadFile(filepath.Join(*audioDir, fileName))
			if err != nil {
				return err
			}

			audioFiles = append(audioFiles, &audioFile{fileName, content})
		}
	}

	if len(audioFiles) == 0 {
		return errNoAudioDir
	}

	logrus.Infof("Loaded %d files", len(audioFiles))

	return nil
}

func handleReq(w http.ResponseWriter, req *http.Request) {
	w.WriteHeader(200)

	go play()
}

func play() {
	if !playing {
		playing = true

		file := audioFiles[rand.Intn(len(audioFiles))]

		logrus.Infof("Playing %s", file.name)

		for _, b := range file.content {
			n, err := otoPlayer.Write([]byte{b})
			if err != nil || n != 1 {
				panic(err)
			}
		}
		playing = false
	}
}
