# gscreen

Serve random images from Google Photos over HTML, intended to be used as a screensaver.

## Usage

- Run `gscreen`, which fetches metadata from Google Photos in the background, and serves images at `http://localhost:9999`:
  ```bash
  $ go build
  $ ./gscreen
  ```
- Point screensaver/browser to `http://localhost:9999` using something like https://github.com/liquidx/webviewscreensaver