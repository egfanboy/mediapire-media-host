module github.com/egfanboy/mediapire-media-host

go 1.16

require (
	github.com/armon/go-metrics v0.4.1 // indirect
	github.com/dhowden/tag v0.0.0-20220618230019-adf36e896086
	github.com/egfanboy/mediapire-common v0.0.0-20220905143518-e3d4d7ef0bac
	github.com/fsnotify/fsnotify v1.5.4
	github.com/google/go-cmp v0.5.8 // indirect
	github.com/google/uuid v1.3.0
	github.com/gorilla/mux v1.8.0
	github.com/hashicorp/consul/api v1.15.3
	github.com/hashicorp/go-cleanhttp v0.5.2 // indirect
	github.com/hashicorp/go-hclog v1.3.1 // indirect
	github.com/hashicorp/go-immutable-radix v1.3.1 // indirect
	github.com/hashicorp/serf v0.10.1 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	github.com/rs/zerolog v1.27.0
	github.com/tcolgate/mp3 v0.0.0-20170426193717-e79c5a46d300
	golang.org/x/net v0.0.0-20220425223048-2871e0cb64e4 // indirect
	golang.org/x/sys v0.1.0 // indirect
	gopkg.in/yaml.v3 v3.0.1
)

//  uncomment for local development
// replace github.com/egfanboy/mediapire-common => ../mediapire-common
// replace github.com/egfanboy/mediapire-manager => ../mediapire-manager
