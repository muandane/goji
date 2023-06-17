# goji

<img align="right" src="examples/go-gopher.gif">

> Commitizen-like tool for formatting commit messages using emojis written in go.

**goji** allows you to easily use emojis in your commits using [git].

```sh
? Select the type of change you are committing: (Use arrow keys)
❯ feature   ✨  Introducing new features.
  fix       🐛  Fixing a bug.
  docs      📚  Documentation change.
  refactor  🎨  Improve structure/format of the code.
  clean     🔥  Remove code or files.
```

## Install

**Homebrew**

```bash
brew tap muandane/gitmoji 
brew install goji
```

**Build locally**

```bash
git clone https://github.com/muandane/goji.git && cd goji
go build goji.go
./goji
```

## Usage

pretty simple just :

```sh
goji
```

![Goji gif](examples/goji-demo.gif)

## Customization

By default `goji` comes ready to run out of the box. *For now customization is on the roadmap (?)*

Uses may vary, so there will be a need for configuration options to allow fine tuning for project needs.
Although if you build locally you can customize the `config.json` file to add or change the scopes, types and other parameters.

## License

Apache 2.0 license [Zine El Abidine Moualhi](https://www.linkedin.com/in/zinemoualhi/)
