# FuzzMux

> FuzzMux (**tmx**) is a generic fuzzy finder to start workspace layouts (or tmux sessions).

Supports multiple option providers, e.g.:

- projects (with customizable checks for git, svn, hg, ...)
- ssh config (parses `~/.ssh/config`)
- kubernetes

The supported window managers / terminal workspace managers are:

- sway
- tmux (only works if no matching option is flagged with `gui: true`)

Feel free to open a PR to add support for your favorite window manager / terminal workspace manager.

## Download

Download the binary from the [GitHub Releases](https://github.com/PhilippHeuer/fuzzmux/releases).

```shell
curl -o ~/.local/bin/tmx -L https://github.com/PhilippHeuer/fuzzmux/releases/latest/download/linux_amd64
```

## Usage

| Command                 | Description                                                         |
|-------------------------|---------------------------------------------------------------------|
| `tmx`                   | Start a layout, lookup over all supported providers                 |
| `tmx project`           | Start a layout for a project                                        |
| `tmx ssh`               | Start a layout for a ssh connection                                 |
| `tmx project -t editor` | Start a layout for a project with a custom layout (bash, nvim, ...) |

## Configure your Providers

```yaml
# yaml-language-server: $schema=https://raw.githubusercontent.com/PhilippHeuer/fuzzmux/main/configschema/v1.json

# project directories
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

# ssh connections
ssh:
  enabled: true

# kubernetes clusters
kubernetes:
  enabled: true
  clusters:
    - name: cluser-1
      tags:
        - production
      kubeconfig: ~/.kube/cluster-1.config
```

## Create your Layouts

The default layout names are identical to the provider names, e.g. `project` for the project provider. This shows how to customize the layout for the `project` provider.

```yaml
layouts:
  project:
    clear-workspace: true # kill all windows in the current workspace before starting
    apps:
      - name: java
        rules:
          - inPath("idea-community") && (contains(TAGS, "language-java") || contains(TAGS, "language-kotlin"))
        commands:
          - command: idea-community "${startDirectory}"
        gui: true
        group: editor # the group is used to ensure only the first matching editor is opened
      - name: vscodium
        rules:
          - inPath("codium") && (contains(TAGS, "language-javascript") || contains(TAGS, "language-typescript"))
        commands:
          - command: codium "${startDirectory}"
        gui: true
        group: editor
      - name: vscode
        rules:
          - inPath("code") && (contains(TAGS, "language-javascript") || contains(TAGS, "language-typescript"))
        commands:
          - command: code ${startDirectory}
        gui: true
        group: editor
      - name: nvim
        rules:
          - inPath("nvim")
        commands:
          - command: nvim +'Telescope find_files hidden=false layout_config={height=0.9}'
        group: editor
```

The command can contain placeholders, e.g.:

- `${name}` - name of the option (for ssh this would be the server alias -> `ssh ${name}`)
- `${display-name}` - display name of the option
- `${start-directory}` - start directory of the option

## Window Manager / Terminal Workspace Manager Setup

Contains information about special setup steps if required.

### tmux

When using tmux, you might want to add the following to your `~/.tmux.conf`:

```bash
# don't exit server without sessions
set -g exit-empty off

# don't exit without attached clients
set -g exit-unattached off

# start at index 1
set -g base-index 1
```

Also ensure you always have a tmux server running (`tmux start-server`, preferably as a user service) to jump into your sessions quickly.

## License

Released under the [MIT license](./LICENSE).
