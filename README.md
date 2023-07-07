[![codecov](https://codecov.io/gh/muandane/goji/branch/main/graph/badge.svg?token=0PYU31AH2S)](https://codecov.io/gh/muandane/goji)
# goji

<img align="right" src="examples/go-gopher.gif">

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
```

## Features

- Interactive CLI for choosing commit types, scopes, and descriptions
- Predefined commit types with corresponding emojis
- Customizable commit types and scopes through a JSON configuration file
- Supports Git out of the box

## Install

**Homebrew**

```bash
brew tap muandane/gitmoji
brew install goji
```

**Build locally**

```bash
git clone https://github.com/muandane/goji.git && cd goji
go build ./src/cmd/goji.go
mv goji /usr/local/bin
goji --version 
```

## Usage

Simply run `goji` in your terminal to start the interactive commit process:

![Goji gif](examples/goji.gif)

## Customization

By default `goji` comes ready to run out of the box and you can initialize a config file with commands. _For now customization is in the works (?)_

```sh
goji --init
```

**HOW TO**

You can customize the `.goji.json` generated file to add or change the scopes, types and other parameters:

```json
{
  "Types": [
    //***
    {
      "emoji": "✨",
      "code": ":sparkles:",
      "description": "Introducing new features.",
      "name": "feature"
    },
    {
      "emoji": "🐛",
      "code": ":bug:",
      "description": "Fixing a bug.",
      "name": "fix"
    }
    //***
  ],
  "Scopes": [
    "home",
    "accounts",
    "ci"
  ],
  "Symbol": true,
  "SkipQuestions": [],
  "SubjectMaxLength": 50
}
```

You can skip questions by adding them in `"SkipQuestions"`

Only `"Scopes"` question can be skipped since it's optional according to the [Commit Spec](https://www.conventionalcommits.org/en/v1.0.0/)

## License

Apache 2.0 license [Zine El Abidine Moualhi](https://www.linkedin.com/in/zinemoualhi/)

## Acknowledgements

Thanks to [@Simplifi-ED](https://www.simplified.fr) & @IT Challenge in letting me work on this open source side project and to my mentor [@EtienneDeneuve](https://github.com/EtienneDeneuve) for the help with learning Go lang.
