[![codecov](https://codecov.io/gh/muandane/goji/branch/main/graph/badge.svg?token=0PYU31AH2S)](https://codecov.io/gh/muandane/goji) [![Go Report Card](https://goreportcard.com/badge/github.com/muandane/goji)](https://goreportcard.com/report/github.com/muandane/goji)

# goji

<img align="right" src="public/go-gopher.gif">

> [!NOTE]
> Commitizen-like tool for formatting commit messages using emojis written in go.

**goji** is an easy-to-use commit message formatting tool, inspired by Commitizen and cz-cli,
that helps you create conventional commits with emojis with streamlined [git] commit process by providing a user-friendly TUI
for selecting the type of change, scope, and description of your commit message..

```sh
? Select the type of change you are committing: (Use arrow keys)
❯ feature   ✨  Introducing new features.
  fix       🐛  Fixing a bug.
  docs      📚  Documentation change.
  refactor  🎨  Improve structure/format of the code.
  clean     🔥  Remove code or files.
? What is the scope of this change? (class or file name): (press [enter] to skip)
>
```

## Features

- Interactive and Non-interactive CLI with sensible defaults for choosing commit types, scopes, and descriptions and signing off commits.
- Predefined commit types with corresponding emojis.
- Customizable commit types and scopes through a JSON configuration file.
- Supports Git out of the box

## Install

**Homebrew (macOs/Linux)**

```sh
brew install muandane/tap/goji
```

**Linux (or WSL)**

```sh
VERSION=$(curl --silent "https://api.github.com/repos/muandane/goji/releases/latest" | jq .tag_name -r)
curl -Lso goji.tar.gz https://github.com/muandane/goji/releases/download/$VERSION/goji_${VERSION}_Linux_x86_64.tar.gz
tar -xvzf goji.tar.gz
chmod +x ./goji
# optionnal
sudo mv ./goji /usr/local/bin/
```

**Windows (winget)**

```sh
winget install muandane.goji
```

**Build locally**

```sh
git clone https://github.com/muandane/goji.git && cd goji
go build -ldflags "-s -w -X goji/cmd.version=0.0.8"
mv goji /usr/local/bin
goji --version
```

**Build with Nix**

```sh
git clone https://github.com/muandane/goji.git && cd goji
nix-build -E 'with import <nixpkgs> {}; callPackage ./default.nix {}'
```

## Usage

Simply run `goji` in your terminal to start the interactive commit process:

![Goji gif](public/goji-demo.gif)

If you don't want the interactive screen you can use the flags to construct a message:

```sh
goji --type feat --scope home --message "Add home page" --git-flag="--porcelain" --git-flag="--branch"  --signoff --no-verify --add 
 
-a, --add                Automatically stage files that have been modified and deleted
--amend                  Change last commit
-n, --no-verify          bypass pre-commit and commit-msg hooks
```

## Check command

To check if a commit message is conventional run:

```sh
goji check
```

To use the check command, add the following to your pre-commit hook in your git repository:

```yaml
- repo: https://github.com/muandane/goji
  rev: v0.0.8
  hooks:
    - id: goji-check
      name: goji check
      description: >
        Check whether the current commit message follows commiting rules. Allow empty commit messages by default, because they typically indicate to Git that the commit should be aborted.
```

## Customization

By default `goji` comes ready to run out of the box and you can initialize a config file with commands. _For now customization is in the works (?)_

```sh
goji init --repo # Writes the config in the git repo's root
goji init --global # Writes the config to home directory
```

**HOW TO**

You can customize the `.goji.json` generated file to add or change the scopes, types and other parameters:

```json
{
  "types": [
    //***
    {
      "emoji": "✨",
      "code": ":sparkles:",
      "description": "Introducing new features.",
      "name": "feat"
    },
    {
      "emoji": "🐛",
      "code": ":bug:",
      "description": "Fixing a bug.",
      "name": "fix"
    }
    //***
  ],
  "scopes": ["home", "accounts", "ci"],
  "noemoji": false,
  "skipquestions": [],
  "subjectmaxlength": 50,
  "signoff": true
}
```

Only `"Scopes"` question can be skipped since it's optional according to the [Commit Spec](https://www.conventionalcommits.org/en/v1.0.0/)

### Configuration options

| Option             | Type             | Description                                                                 |
| ------------------ | ---------------- | --------------------------------------------------------------------------- |
| `types`            | Array of objects | Types for the commit messages (emoji, code, description, name)              |
| `scopes`           | Array of strings | Optional scopes for the commit messages (you can auto-complete with ctrl+e) |
| `noemoji`          | Boolean          | Creates commit message with emojis in types                                 |
| `subjectmaxlength` | Number           | Maximum length for the description message                                  |
| `signoff`          | Boolean          | Add a sign off to the end of the commit message                             |
| `skipquestions`    | Array of strings | Skip prompting for these questions (Unimplemented)                          |

## License

Apache 2.0 license [Zine El Abidine Moualhi](https://www.linkedin.com/in/zinemoualhi/)

## Acknowledgements

Thanks to [@Simplifi-ED](https://www.simplified.fr) & @IT Challenge in letting me work on this open source side project and to my mentor [@EtienneDeneuve](https://github.com/EtienneDeneuve) for the help with learning Go lang.

<img align="center" src="public/logo.svg"  alt="IT Challenge" width="200"/>
