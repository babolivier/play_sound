This project is a small experiment I made, of a very simple program that starts a web
server and plays a sound for each request it gets on `/play`. It's built with
[oto](https://github.com/hajimehoshi/oto). For a bigger and more viable project you'll
probably want something more complete like [beep](https://github.com/faiface/beep).

It accepts a few command-line arguments and flags:

* `-f` (required): the path to the audio file to use; must be in a lossless format (WAV is
  preferred).
* `-p` (default: `8888`): the port the web server will listen to.
* `-s` (default: `48000`): the audio file's sample rate (in Hz).
* `-m` (flag; optional): use this flag if the file only uses one channel (mono).