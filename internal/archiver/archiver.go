package archiver

type Archiver interface {
	Archive(source, destination string) error
}
