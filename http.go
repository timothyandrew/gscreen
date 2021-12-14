package main

import (
	"fmt"
	"log"
	"net/http"
)

func image(cache *MetadataCache, client *http.Client) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		item := cache.Random()

		url, err := GetImageUrl(client, item)
		if err != nil {
			fmt.Fprintln(w, "ERROR", err)
		}

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
						}, 8000);
						
						console.debug("Cached media items: ", %d)
					</script>
				</body>
			</html>
		`

		fmt.Fprintf(w, s, url, len(cache.cache))
	}
}

func InitializeHttpServer(cache *MetadataCache, client *http.Client) {
	http.HandleFunc("/", image(cache, client))
	log.Println("Starting HTTP server at :9999")
	http.ListenAndServe(":9999", nil)
}
