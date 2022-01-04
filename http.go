package main

import (
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
)

func html(cache DownloadCache) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		cached := <-cache
		imageStr := base64.StdEncoding.EncodeToString(cached.data)

		mediaHtml := fmt.Sprintf(`<img src="data:image/jpeg;charset=utf-8;base64,%s" />`, imageStr)

		if cached.mediaItem.isVideo() {
			mediaHtml = fmt.Sprintf(`
			<video autoplay muted loop>
				<source type="%s" src="data:%s;base64,%s" />
			</video>
			`, cached.mediaItem.MimeType, cached.mediaItem.MimeType, imageStr)
		}

		s := `
			<html>
				<head>
					<style>
						div {
							position: absolute;
							height: 100vh;
							width: 100%%;
							top: 0;
							left: 0;
							margin: 0;
							padding: 0;
						}
						img, video {
							margin: 0 auto;
							height: 100%%;
							display: block;
						}
						span {
							color: #fff;
							margin: 10px auto;
							font-family: ui-sans-serif, system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, "Noto Sans", sans-serif, "Apple Color Emoji", "Segoe UI Emoji", "Segoe UI Symbol", "Noto Color Emoji";
							text-align: center;
							width: 100%%;
							display: block;
						}
					</style>
				</head>
				<body>
					<div>
						<span>%s</span>
						%s
					</div>
					<script type="text/javascript">
						window.setTimeout(function() {
							window.location.reload();
						}, 5000);
					</script>
				</body>
			</html>
		`

		fmt.Fprintf(w, s, cached.mediaItem.Metadata.CreationTime, mediaHtml)
	}
}

func InitializeHttpServer(cache DownloadCache) {
	http.HandleFunc("/", html(cache))
	log.Println("Starting HTTP server at :9999")
	http.ListenAndServe("0.0.0.0:9999", nil)
}
