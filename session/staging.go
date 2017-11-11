package session

import (
	"hash"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/pkg/errors"

	"github.com/havoc-io/mutagen/sync"
)

const (
	// numberOfByteValues is the number of values a byte can take.
	numberOfByteValues = 1 << 8
)

type stagingSink struct {
	coordinator *stagingCoordinator
	// path is the path that is being staged. It is not the path to the storage
	// or the staging destination.
	path     string
	storage  *os.File
	digester hash.Hash
}

func (s *stagingSink) Write(data []byte) (int, error) {
	// Write to the underlying storage.
	n, err := s.storage.Write(data)

	// Write as much to the digester as we wrote to the underlying storage. This
	// can't fail.
	s.digester.Write(data[:n])

	// Done.
	return n, err
}

func (s *stagingSink) Close() error {
	// Close the underlying storage.
	if err := s.storage.Close(); err != nil {
		return errors.Wrap(err, "unable to close underlying storage")
	}

	// Compute the final digest.
	digest := s.digester.Sum(nil)

	// Compute where the file should be relocated.
	destination, prefix, err := pathForStaging(s.coordinator.root, s.path, digest)
	if err != nil {
		os.Remove(s.storage.Name())
		return errors.Wrap(err, "unable to compute staging destination")
	}

	// Ensure the prefix directory exists.
	if err = s.coordinator.ensurePrefixExists(prefix); err != nil {
		os.Remove(s.storage.Name())
		return errors.Wrap(err, "unable to create prefix directory")
	}

	// Relocate the file to the destination.
	if err = os.Rename(s.storage.Name(), destination); err != nil {
		os.Remove(s.storage.Name())
		return errors.Wrap(err, "unable to relocate file")
	}

	// Success.
	return nil
}

// stagingCoordinator coordinates the reception of files via rsync (by
// implementing rsync.Sinker) and the provision of those files to transitions
// (by implementing sync.Provider). It is not safe for concurrent access, and
// each stagingSink it produces should be closed before another is created.
type stagingCoordinator struct {
	// version is the session version.
	version Version
	// root is the staging root.
	root string
	// rootCreated indicates whether or not the staging root has been created
	// by us since the last wipe.
	rootCreated bool
	// prefixCreated maps prefix names (e.g. "00" - "ff") to a boolean
	// indicating whether or not the prefix has been created by us since the
	// last wipe.
	prefixCreated map[string]bool
}

func newStagingCoordinator(session string, version Version, alpha bool) (*stagingCoordinator, error) {
	// Compute the staging root.
	root, err := pathForStagingRoot(session, alpha)
	if err != nil {
		return nil, errors.Wrap(err, "unable to compute staging root")
	}

	// Success.
	return &stagingCoordinator{
		version:       version,
		root:          root,
		prefixCreated: make(map[string]bool, numberOfByteValues),
	}, nil
}

func (c *stagingCoordinator) ensurePrefixExists(prefix string) error {
	// Check if we've already created that prefix.
	if c.prefixCreated[prefix] {
		return nil
	}

	// Otherwise create it and mark it as created. We can also mark the root as
	// created since it'll be an intermediate directory.
	if err := os.MkdirAll(filepath.Join(c.root, prefix), 0700); err != nil {
		return err
	}
	c.rootCreated = true
	c.prefixCreated[prefix] = true

	// Success.
	return nil
}

func (c *stagingCoordinator) wipe() error {
	// Reset the prefix creation tracker.
	c.prefixCreated = make(map[string]bool, numberOfByteValues)

	// Reset root creation tracking.
	c.rootCreated = false

	// Remove the staging root.
	if err := os.RemoveAll(c.root); err != nil {
		errors.Wrap(err, "unable to remove staging directory")
	}

	// Success.
	return nil
}

func (c *stagingCoordinator) Sink(path string) (io.WriteCloser, error) {
	// Create the staging root if we haven't already.
	if !c.rootCreated {
		if err := os.MkdirAll(c.root, 0700); err != nil {
			return nil, errors.Wrap(err, "unable to create staging root")
		}
		c.rootCreated = true
	}

	// Create a temporary storage file in the staging root.
	storage, err := ioutil.TempFile(c.root, "staging")
	if err != nil {
		return nil, errors.Wrap(err, "unable to create temporary storage file")
	}

	// Success.
	return &stagingSink{
		coordinator: c,
		path:        path,
		storage:     storage,
		digester:    c.version.hasher(),
	}, nil
}

func (c *stagingCoordinator) Provide(path string, entry *sync.Entry) (string, error) {
	// Compute the expected location of the file.
	expectedLocation, _, err := pathForStaging(c.root, path, entry.Digest)
	if err != nil {
		return "", errors.Wrap(err, "unable to compute staging path")
	}

	// Ensure that it has the correct permissions. This will fail if the file
	// doens't exist.
	permissions := os.FileMode(0600)
	if entry.Executable {
		permissions = os.FileMode(0700)
	}
	if err = os.Chmod(expectedLocation, permissions); err != nil {
		return "", errors.Wrap(err, "unable to set file permissions")
	}

	// Success.
	return expectedLocation, nil
}
