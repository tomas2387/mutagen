# Mutagen

Mutagen is a **fast**, continuous, bidirectional file synchronization tool. It
can safely, scalably, and efficiently synchronize filesystem contents between
arbitrary pairs of locations in near real-time. Support is currently implemented
for locations on local filesystems, SSH-accessible filesystems, and Docker
container filesystems.

Mutagen excels at supporting remote development scenarios, with options
specifically designed to help developers edit code locally while building,
running, or packaging it in a remote environment.

For a basic summary and quick usage information, please check out the
[documentation site](https://mutagen.io).

Detailed information about Mutagen's [unique features](#unique-features) and
[advanced usage](#usage) can be found below.

To follow release and security announcements for Mutagen, please subscribe to
the [mutagen-announce](https://groups.google.com/forum/#!forum/mutagen-announce)
mailing list.

For discussion about Mutagen usage, please join the
[discussion forums](https://groups.google.com/forum/#!forum/mutagen).


## Status

Mutagen is a very powerful tool that is still in early beta. It will almost
certainly have unknown issues. It should not be used on production or
mission-critical systems. Use on *any* system is at your own risk (please see
the [license](https://github.com/havoc-io/mutagen/blob/master/LICENSE)).

That being said, Mutagen is a very useful tool and I use it daily for work on
remote systems. The more people who use it and report
[issues](https://github.com/havoc-io/mutagen/issues), the better it will get!

| Windows                           | macOS/Linux                                   | Code coverage                           | Report card                           |
| :-------------------------------: | :-------------------------------------------: | :-------------------------------------: | :-----------------------------------: |
| [![Windows][win-badge]][win-link] | [![macOS/Linux][mac-lin-badge]][mac-lin-link] | [![Code coverage][cov-badge]][cov-link] | [![Report card][rc-badge]][rc-link]   |

[win-badge]: https://ci.appveyor.com/api/projects/status/qywidv5a1vf7g3b5/branch/master?svg=true "Windows build status"
[win-link]:  https://ci.appveyor.com/project/havoc-io/mutagen/branch/master "Windows build status"
[mac-lin-badge]: https://travis-ci.org/havoc-io/mutagen.svg?branch=master "macOS/Linux build status"
[mac-lin-link]:  https://travis-ci.org/havoc-io/mutagen "macOS/Linux build status"
[cov-badge]: https://codecov.io/gh/havoc-io/mutagen/branch/master/graph/badge.svg "Code coverage status"
[cov-link]: https://codecov.io/gh/havoc-io/mutagen/tree/master/pkg "Code coverage status"
[rc-badge]: https://goreportcard.com/badge/github.com/havoc-io/mutagen "Report card status"
[rc-link]: https://goreportcard.com/report/github.com/havoc-io/mutagen "Report card status"


## Security

Mutagen takes security very seriously. If you believe you have found a security
issue with Mutagen, please practice responsible disclosure practices and send an
email directly to [security@mutagen.io](mailto:security@mutagen.io) instead of
opening a GitHub issue. For more information, please see the
[security documentation](SECURITY.md).


## Usage

**For a quick summary and usage guide that will cover most of your needs, please
see the [documentation site](https://mutagen.io).**

For information about Mutagen's SSH support, please see the
[SSH documentation](doc/ssh.md).

For information about Mutagen's Docker support, please see the
[Docker documentation](doc/docker.md).

For information about Mutagen's configuration system, please see the
[configuration documentation](doc/configuration.md).

For information about using Mutagen with version control systems, please see the
[VCS documentation](doc/vcs.md).

For information about symlink support, please see the
[symlink documentation](doc/symlinks.md).

For information about ignoring files, please see the
[ignore documentation](doc/ignores.md).

For information about filesystem watching, please see the
[watching documentation](doc/watching.md).

For information about Mutagen's safety mechanisms, please see the
[safety documentation](doc/safety.md).

For platform-specific instructions and known issues, please see the
[platform guide](doc/platforms.md).


## Contributing

If you'd like to contribute to Mutagen, please see the
[contribution documentation](CONTRIBUTING.md).


## Unique features

Instead of providing a heavily biased feature comparison table, I'll just point
out what I consider to be the unique and compelling features of Mutagen. Astute
readers with knowledge of the file synchronization landscape can draw their own
conclusions. I'd recommend that users read this list so they know what they're
getting.

- Mutagen is truly cross-platform, treating Linux, macOS, Windows, and other
  operating systems as first class citizens. Differences in OS and filesystem
  behavior are addressed head-on, not ignored until an edge case causes
  breakage. Mutagen attempts to handle quirks by default, e.g. dealing with
  case-(in)sensitivity, HFS's pseudo-NFD Unicode normalization, filesystems that
  don't support POSIX executability bits, or file names that might create NTFS
  alternate data streams.
- Mutagen is a user-space utility, not requiring any kernel extensions or
  administrative permissions to use.
- **Mutagen only needs to be installed on the computer where you want to control
  synchronization.** Mutagen comes with a broad range of small, cross-compiled
  "agent" binaries that it automatically copies to remote endpoints as
  necessary. Most major platforms and architectures are supported.
- Mutagen is designed to handle very large directory hierarchies efficiently. It
  maintains a filesystem cache to allow quick re-scans and uses the
  [rsync algorithm](https://rsync.samba.org/tech_report/) to transfer filesystem
  scans and files themselves. File transfers are also pipelined to mitigate the
  effects of latency. Mutagen won't break a sweat on a GB-sized directory
  hierarchy containing 100,000 files.
- Mutagen propagates changes bidirectionally. Any conflicts that arise will be
  flagged for resolution. Automatic conflict resolution is performed if doing so
  does not result in the destruction of unsynchronized data. Manual conflict
  resolution is performed by manually deleting the undesired side of the
  conflict. Conflicts won't stop non-conflicting changes from propagating.
- Mutagen is robust to connection drop-outs. It will attempt to reconnect
  automatically to endpoints and will resume synchronization safely. In the mean
  time, your local copy of a synchronization root continues to exist on the
  filesystem for you to access and edit like any other files. Once
  synchronization resumes, Mutagen will continue right where it left off, even
  resuming partially completed file staging.
- Mutagen identifies changes to file contents rather than just modification
  times.
- On systems that support recursive filesystem watching (macOS and Windows),
  Mutagen effeciently watches synchronization roots for changes. Other systems
  currently use regular and efficient polling out of a desire to support very
  large directory hierarchies that might exhaust watch and file descriptors. On
  Linux, Mutagen couples this polling with a restricted set of native watches on
  the most recently updated contents in order to maintain low-latency change
  notifications.
- Mutagen is agnostic of the transport to endpoints - all it requires is a byte
  stream to each endpoint. Support is currently implemented for synchronization
  with local, SSH, and Docker endpoints, but support for other remote types can
  easily be added. As a corollary, Mutagen can even synchronize between two
  remote endpoints without ever needing a local copy of the files.
- Mutagen can display dynamic synchronization status in the terminal.
- Mutagen does not propagate (most) permissions, but it does
  preserve<sup>1</sup> permissions when updating files. The only permission
  propagated by Mutagen is executability or lack thereof. Any other permissions
  are left untouched for existing files and set to user-only access for newly
  created files. This is by design, since Mutagen's main purpose is remote
  development. Nothing in the current design precludes adding more extensive
  permission propagation in the future.
- Mutagen has (**best-effort**) [safety mechanisms](doc/safety.md) to avoid
  accidental data loss.

<sup>1</sup> This preservation behavior is currently limited to POSIX systems,
but should be coming to Windows systems soon.

You might have guessed that Mutagen's closest cousin is the
[Unison](http://www.cis.upenn.edu/~bcpierce/unison) file synchronization tool.
This tool has existed for ages, and while it is *very* good at what it does, it
didn't quite fit my needs. In particular, it has a *lot* of knobs to turn, puts
a lot of focus on transferring permissions (which can cause even more
headaches), and requires installation on both ends of the connection. I wanted
something simpler, a bit more performant, and just a bit more modern (the fact
that Unison is written in rather terse OCaml also makes it a bit difficult to
extend or support on more obscure platforms and architectures).


## Building

Please see the [build instructions](doc/building.md).
