This project is a small experiment I made, of a very simple program that starts a web
server and plays a random sound from a directory of audio files for each request it gets
on `/play`. It's built with [oto](https://github.com/hajimehoshi/oto). For a bigger and
more viable project you'll probably want something more complete like
[beep](https://github.com/faiface/beep) (which also uses oto behind the scenes).

It accepts a few command-line arguments and flags:

* `-d` (required): the path to the directory containing the audio files to use; the files
  must be in the WAV format.
* `-p` (default: `8888`): the port the web server will listen to.
* `-s` (default: `48000`): the audio file's sample rate (in Hz).
* `-m` (flag; optional): use this flag if the file only uses one channel (mono).