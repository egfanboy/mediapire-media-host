module github.com/egfanboy/mediapire-media-host

go 1.16

require (
	github.com/dhowden/tag v0.0.0-20220618230019-adf36e896086
	github.com/egfanboy/mediapire-common v0.0.0-20220828193527-bd4d006a2cb6
	github.com/fsnotify/fsnotify v1.5.4
	github.com/gorilla/mux v1.8.0
	github.com/rs/zerolog v1.27.0
	github.com/tcolgate/mp3 v0.0.0-20170426193717-e79c5a46d300
	gopkg.in/yaml.v3 v3.0.1
)

//  uncomment for local development
// replace github.com/egfanboy/mediapire-common => ../mediapire-common
