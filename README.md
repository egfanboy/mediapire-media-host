## Mediapire

`Mediapire` is a set of services to help manage and access your media across different machines. It consists of a [management plane](https://github.com/egfanboy/mediapire-manager) that is tasked to manage and communicate with the media hosts that contain your media.

## Mediapire Media Host

The media host service is a part of the `Mediapire` suite. It will scan the media in the configured directories based on the file types that you configured to track and expose APIs that will allow the management plane to access the media of the current host.

### Getting started

1. Create your config file

```bash
cp config.example.yaml config.yaml
```

2. Enter the proper values in your config file

```yml
# Directories to scan for media
# Note, this will scan directories recursively
directories:
  - directory1
  - directory2
# Type of media we want to scan
fileTypes:
  - mp3
  - mp4
# configuration for media host (self)
mediaHost:
  scheme: http
  # port of the media host (self) instance
  port: 444
  # connection information for media manager
manager:
  scheme: http
  host: "192.168.4.98"
  port: 9999
```

**Note** The manager information is for the instance of the [mediapire manager](https://github.com/egfanboy/mediapire-manager) that you need to have running.

3. Run the media host

On startup, the media host will scan the media based on the provided file extensions found in the directories provided and attempt to register itself to the manager. Therefore, the manager must be already be running ([more on manager here](https://github.com/egfanboy/mediapire-manager)).

### Directory Scanning

The mediapire media host will scan the files in the configured directories and only track the files of the configured file types. Filesystem watchers will then be created that will rescan the media in the directories once the media in any of the directories is changed (Create, update, delete).
