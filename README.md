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

- backstage (query catalog)
- jira (query issues)
- keycloak (query users, groups and clients)
- kubernetes clusters (including openshift)
- ldap (query users and groups)
- projects (with depth and customizable checks for git, svn, hg, ...)
- rundeck (query jobs)
- ssh config (parses `~/.ssh/config`)
- usql connections (parses `~/.config/usql/config.yml`)
- firefox bookmarks (specify sqlite db file to query)
- desktop applications (parses XDG `.desktop` entries, e.g. on Ubuntu, Fedora, NixOS, Arch)

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

## Configure Modules

### Backstage

The `backstage` module can query components in the catalog.

```yaml
modules:
  - type: backstage
    host: https://demo.backstage.io
    bearer-token: secret # optional, see https://backstage.io/docs/auth/service-to-service-auth/#static-tokens
    attribute-mapping:
      - source: metadata.name
        target: name
    query:
      - service
      - openapi
      - user
      - website
      - team
      - department
```

### JIRA

The `jira` module can query issues from JIRA.

```yaml
modules:
  - type: jira
    host: https://issues.apache.org/jira/
    bearer-token: secret # your personal access token
    jql: project = AMQ # optional filter, see https://support.atlassian.com/jira-software-cloud/docs/jql-fields/
    attribute-mapper:
      - source: project
        target: project
      - source: summary
        target: summary
      - source: type
        target: type
      - source: status
        target: status
      - source: priority
        target: priority
      - source: assignee
        target: assignee
      - source: reporter
        target: reporter
```

### Keycloak

The `keycloak` module can query users, groups and clients across all realms the user has access to.

```yaml
modules:
  - type: keycloak
    host: http://localhost:8080
    realm: master
    username: admin
    password: secret
    query:
      - user
      - client
      - group
```

### Kubernetes

The `kubernetes` module supports multiple clusters and can query namespaces.

```yaml
modules:
  - type: kubernetes
    clusters:
      - name: cluster01
        tags:
          - production
        kubeconfig: ~/.kube/cluster.config
```

### LDAP

The `ldap` module can query users and groups from LDAP or Active Directory.

```yaml
modules:
  - name: ldap-users
    type: ldap
    host: ldap://127.0.0.1:389
    base-dn: "dc=company,dc=com"
    bind-dn: "cn=admin,dc=company,dc=com"
    bind-password: "secret"
    filter: (&(objectClass=organizationalPerson))
  - name: ldap-groups
    type: ldap
    host: ldap://127.0.0.1:389
    base-dn: "dc=company,dc=com"
    bind-dn: "cn=admin,dc=company,dc=com"
    bind-password: "secret"
    filter: (|(objectClass=group)(objectClass=posixGroup)(objectClass=groupOfNames))
```

### Project

The `project` module can query projects from your local filesystem.

```yaml
modules:
  - type: project
    display-format: relative
    directories:
      - path: ~/projects/Golang
        depth: 1
      - path: ~/projects/Rust
        depth: 1
      - path: ~/projects/Java
        depth: 1
```

### Rundeck

The `rundeck` module can query jobs from the rundeck job scheduler.

```yaml
modules:
  - type: rundeck
    host: http://localhost:4440
    token: your-personal-access-token
    projects:
      - test
```

### SSH

The `ssh` module reads connections from the `~/.ssh/config` file.

```yaml
modules:
  - type: ssh
    start-directory: "~"
```

### USQL

The `usql` module reads db connections from the `~/.config/usql/config.yml` file.

```yaml
modules:
  - type: usql
    start-directory: "~"
```

### Browsers

The `firefox` module can query bookmarks from Firefox-based browsers (e.g. Firefox, Librewolf, Floorp, Zen):

```yaml
- type: firefox
  name: firefox
  profile-path: ~/.mozilla/firefox/abcd1234.default
- type: firefox
  name: librewolf
  profile-path: ~/.librewolf/abcd1234.default
- type: firefox
  name: floorp
  profile-path: ~/.floorp/abcd1234.default
- type: firefox
  name: zen
  profile-path: ~/.zen/abcd1234.default
```

### Desktop Applications

The `desktop` module discovers installed applications from XDG desktop entries in `$XDG_DATA_HOME/applications` and `$XDG_DATA_DIRS/*/applications`.
This covers distribution packages, Flatpak, Snap and Nix profiles, as long as their directories are part of `XDG_DATA_DIRS`.
User entries take precedence over system entries, and entries flagged as `Hidden`, `NoDisplay` (unless `show-hidden` is set), restricted by `OnlyShowIn`/`NotShowIn` or with a missing `TryExec` binary are skipped.

```yaml
modules:
  - type: desktop
    # optional: additional directories to search (lowest precedence)
    directories:
      - ~/.local/share/flatpak/exports/share/applications
    # optional: only include / exclude specific categories
    categories: []
    exclude-categories:
      - Settings
    # optional: exclude specific desktop file ids
    exclude:
      - org.gnome.Settings
    # optional: include entries flagged with NoDisplay=true
    show-hidden: false
```

The built-in `desktop` layout starts graphical applications directly through the window manager (`hyprland`, `sway`, `i3`, `gnome`), and runs `Terminal=true` applications in a terminal / tmux window.
The command is available as `{{!exec}}` (already shell-quoted), alongside `{{desktopId}}`, `{{desktopFile}}`, `{{genericName}}`, `{{icon}}`, `{{categories}}` and `{{keywords}}`.

### Static

The `static` module can be used to define static options.

```yaml
modules:
  - id: my-static-option
    name: "My Static Option"
    display-name: "My Static Option"
    description: "This is a static option"
    web: "https://example.com"
    tags:
      - static
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
          - command: idea-community "{{startDirectory}}"
        gui: true
        group: editor # the group is used to ensure only the first matching editor is opened
      - name: vscodium
        rules:
          - inPath("codium") && (contains(TAGS, "language-javascript") || contains(TAGS, "language-typescript"))
        commands:
          - command: codium "{{startDirectory}}"
        gui: true
        group: editor
      - name: vscode
        rules:
          - inPath("code") && (contains(TAGS, "language-javascript") || contains(TAGS, "language-typescript"))
        commands:
          - command: code "{{startDirectory}}"
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

- `{{name}}` - name of the option (for ssh this would be the server alias -> `ssh "{{name}}"`)
- `{{displayName}}` - display name of the option
- `{{startDirectory}}` - start directory of the option
- `{{id}}`, `{{description}}`, `{{web}}`, `{{providerName}}`, `{{providerType}}` - other option fields
- any key of the option context, e.g. `{{namespace}}` or `{{kubeConfig}}` for kubernetes

`{{key}}` escapes the value for use inside double quotes (`"{{name}}"`), use `{{!key}}` to insert the raw value.
Environment variables (`$HOME`, `${USER}`) in the command are expanded before placeholders are replaced.

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
