# FuzzMux

> FuzzMux (**tmx**) is a generic fuzzy finder to start tmux sessions with pre-defined layouts / init commands.

Supports multiple providers to find sessions to start, e.g.:

- projects (with customizable checks for git, svn, hg, ...)
- ssh config (parses `~/.ssh/config`)

## Download

Download the binary from the [GitHub Releases](https://github.com/PhilippHeuer/fuzzmux/releases).

## tmux

- ensure the tmux server is running (`tmux start-server`)
- add the following to your `~/.tmux.conf`:

```bash
# don't exit server without sessions
set -g exit-empty off

# don't exit without attached clients
set -g exit-unattached off
```

## config

### schema support

Add the following to your config to enable schema support:

```yaml
# yaml-language-server: $schema=https://raw.githubusercontent.com/PhilippHeuer/fuzzmux/main/configschema/v1.json
```

### project provider

```yaml
project:
  enabled: true
  display-format: rel
  directories:
    - path: ~/projects/Golang
      depth: 1
    - path: ~/projects/Rust
      depth: 1
    - path: ~/projects/Java
      depth: 1
```

### ssh provider

```yaml
ssh:
  enabled: true
```

## Usage

| Command                 | Description                                                                   |
|-------------------------|-------------------------------------------------------------------------------|
| `tmx`                   | Start a new tmux session, lookup over all supported providers                 |
| `tmx project`           | Start a new tmux session for a project                                        |
| `tmx ssh`               | Start a new tmux session for a ssh connection                                 |
| `tmx project -t editor` | Start a new tmux session for a project with a custom layout (bash, nvim, ...) |

**Example Template:**

```yaml
# templates - can also overwrite the default templates (names: default, project, ssh)
window-template:
  editor: # template name
  - id: 1
    name: bash
  - id: 2
    name: nvim
    commands:
    - nvim +'Telescope find_files hidden=false layout_config={height=0.9}' # open nvim with telescope
    default: true # will select this window by default
```

The command can contain placeholders, e.g.:

- `${name}` - name of the option (for ssh this would be the server alias -> `ssh ${name}`)
- `${display-name}` - display name of the option
- `${start-directory}` - start directory of the option

## License

Released under the [MIT license](./LICENSE).
