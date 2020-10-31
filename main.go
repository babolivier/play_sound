package main

import (
	"errors"
	"flag"
	"io/ioutil"
	"net/http"

	"github.com/hajimehoshi/oto"
)

var (
	port       = flag.String("p", "8888", "Port to listen on")
	audioFile  = flag.String("f", "", "File to play")
	sampleRate = flag.Int("r", 48000, "Sample rate")
	mono       = flag.Bool("m", false, "Use mono instead of stereo")

	otoPlayer   *oto.Player
	fileContent []byte
	playing     bool
)

func main() {
	flag.Parse()

	if *audioFile == "" {
		panic(errors.New("please provide a file with -f"))
	}

	var err error
	fileContent, err = ioutil.ReadFile(*audioFile)
	if err != nil {
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

func handleReq(w http.ResponseWriter, req *http.Request) {
	w.WriteHeader(200)

	if !playing {
		playing = true
		for _, b := range fileContent {
			n, err := otoPlayer.Write([]byte{b})
			if err != nil || n != 1 {
				panic(err)
			}
		}
		playing = false
	}
}
