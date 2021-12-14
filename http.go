package main

import (
	"fmt"
	"log"
	"net/http"
)

func image(cache *MetadataCache) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		item := cache.Random()

		s := `
			<html>
				<head>
					<style>
						img {
							height: 100%%;
							margin: 0 auto;
							display: block;
						}
					</style>
				</head>
				<body>
					<img src="%s=w4000" />
					<script type="text/javascript">
						window.setTimeout(function() {
							window.location.reload();
						}, 5000);
					</script>
				</body>
			</html>
		`

		fmt.Fprintf(w, s, item.BaseUrl)
	}
}

func InitializeHttpServer(cache *MetadataCache) {
	http.HandleFunc("/", image(cache))
	log.Println("Starting HTTP server at :9999")
	http.ListenAndServe(":9999", nil)
}
