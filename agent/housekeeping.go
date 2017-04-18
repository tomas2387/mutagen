package agent

import (
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/shibukawa/extstat"

	"github.com/havoc-io/mutagen/filesystem"
	"github.com/havoc-io/mutagen/process"
)

const (
	maximumAgentIdlePeriod = 30 * 24 * time.Hour
)

func Housekeep() {
	// Compute the path to the agents directory. If we fail, just abort. We
	// don't attempt to create the directory, because if it doesn't exist, then
	// we don't need to do anything and we'll just bail when we fail to list the
	// agent directory below.
	agentsDirectoryPath, err := filesystem.Mutagen(false, agentsDirectoryName)
	if err != nil {
		return
	}

	// Get the list of locally installed agent versions. If we fail, just abort.
	agentVersions, err := filesystem.DirectoryContents(agentsDirectoryPath)
	if err != nil {
		return
	}

	// Compute the name of the agent binary.
	agentName := process.ExecutableName(agentBaseName, runtime.GOOS)

	// Grab the current time.
	now := time.Now()

	// Loop through each agent version, compute the time it was last launched,
	// and remove it if longer than the maximum allowed period. Ignore any
	// failures.
	for _, v := range agentVersions {
		if stat, err := extstat.NewFromFileName(filepath.Join(agentsDirectoryPath, v, agentName)); err != nil {
			continue
		} else if now.Sub(stat.AccessTime) > maximumAgentIdlePeriod {
			os.RemoveAll(filepath.Join(agentsDirectoryPath, v))
		}
	}
}
