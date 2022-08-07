package fs

type FsApi interface {
	WatchDirectory(directory string) error
	CloseWatchers()
}
