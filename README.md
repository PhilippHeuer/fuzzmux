# FuzzMux

[![Go Report Card](https://goreportcard.com/badge/github.com/PhilippHeuer/fuzzmux)](https://goreportcard.com/report/github.com/PhilippHeuer/fuzzmux)
[![OpenSSF Scorecard](https://api.securityscorecards.dev/projects/github.com/PhilippHeuer/fuzzmux/badge)](https://securityscorecards.dev/viewer/?uri=github.com/PhilippHeuer/fuzzmux)

## What is FuzzMux?

FuzzMux provides a global fuzzy finder menu for your projects, ssh connections, namespaces in kubernetes, ... to quickly switch between contexts.
Additionally, it provides a way to define layouts / rules, to start specific applications based on the selected item.

For example:

- open `IntelliJ IDEA` if the project is a `Java project`
- open `Goland` if the project is a `Go project`
- open `Neovim` for all other projects

Similar rules can be defined for ssh connections, kubernetes namespaces, ... - see the [examples/config.yaml](examples/config.yaml).

## Supported Providers

- projects (with depth and customizable checks for git, svn, hg, ...)
- ssh config (parses `~/.ssh/config`)
- kubernetes clusters (including openshift)
- usql connections (parses `~/.config/usql/config.yml`)

## Supported Window Managers / Terminal Workspace Managers

- hyprland
- sway
- i3
- tmux
- shell (fallback, runs the default command in the current shell)

**Note:**: `tmux` and `shell` will ignore options flagged as `gui: true`.

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
| `tmx menu`              | Interactive menu to choose a provider, and then an option           |

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

Specific setup steps, if required.

#### tmux

- Ensure you have a tmux server running (`tmux start-server`, preferably as a user service) to jump into your sessions quickly.
- Add the options in [tmux.conf](examples/tmux.conf) to your `~/.tmux.conf`.

## Credits

- [junegunn/fzf](https://github.com/junegunn/fzf) - A command-line fuzzy finder
- [ktr0731/go-fuzzyfinder](https://github.com/ktr0731/go-fuzzyfinder) - A fuzzy finder for Go
- [cidverse/repoanalyzer](https://github.com/cidverse/repoanalyzer) - Scans your repositories to get insights about the languages, build systems, and more for your rules

## Contributing

Contributions are welcome!

## License

Released under the [MIT license](./LICENSE).
