package main

import (
	"fmt"
	"log"
	"net/http"
)

func image(cache DownloadCache) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		w.Write(<-cache)
	}
}

func html(w http.ResponseWriter, req *http.Request) {
	s := `
			<html>
				<head>
					<style>
						img {
							height: 100%;
							margin: 0 auto;
							display: block;
						}
					</style>
				</head>
				<body>
					<img src="/image" />
					<script type="text/javascript">
						window.setTimeout(function() {
							window.location.reload();
						}, 5000);
					</script>
				</body>
			</html>
		`

	fmt.Fprint(w, s)
}

func InitializeHttpServer(cache DownloadCache) {
	http.HandleFunc("/", html)
	http.HandleFunc("/image", image(cache))
	log.Println("Starting HTTP server at :9999")
	http.ListenAndServe(":9999", nil)
}
