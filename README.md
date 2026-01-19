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
- AI generated commit messages using multiple providers:
  - **OpenRouter** (default, access to multiple models, requires API key)
  - **Groq** (fast inference, requires API key)
- Intelligent large diff handling with automatic summarization
- No rate limiting issues with smart diff compression

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
# Auto-detect version from git tags (recommended)
VERSION=$(git describe --tags --always --dirty | sed 's/^v//')
go build -ldflags "-s -w -X github.com/muandane/goji/cmd.version=${VERSION}"
mv goji /usr/local/bin
goji --version
```

Or manually specify the version:

```sh
go build -ldflags "-s -w -X github.com/muandane/goji/cmd.version=0.1.8"
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

## Draft command

The `draft` command allows you to generate a commit message for staged changes using various AI providers.

```sh
goji draft
```

This command connects to an AI provider to generate a commit message based on your staged changes.

### Quick Start

**1. Set up environment variables:**

```sh
# For OpenRouter (default)
export OPENROUTER_API_KEY="your-openrouter-api-key"

# For Groq
export GROQ_API_KEY="your-groq-api-key"
```

**2. Generate and commit:**

```sh
git add .
goji draft --commit
```

### Basic Usage

```sh
# Generate a commit message (interactive)
goji draft

# Generate and commit directly
goji draft --commit

# Generate with detailed body
goji draft --body

# Generate and commit with detailed body
goji draft --body --commit
```

### Command Options

- `-c`, `--commit`: Commit the generated message directly.
- `-b`, `--body`: Generate a detailed commit message with body.
- `-t`, `--type`: Override the commit type (e.g., feat, fix, docs).
- `-s`, `--scope`: Override the commit scope (e.g., api, ui, core).

### AI Providers

Goji supports multiple AI providers for generating commit messages. Configure your preferred provider in the `.goji.json` file:

#### 1. OpenRouter (Default)

Access to multiple AI models through a single API. Requires an OpenRouter API key.

**Setup:**

```sh
export OPENROUTER_API_KEY="your-openrouter-api-key"
```

**Configuration:**

```json
{
  "aiprovider": "openrouter",
  "aichoices": {
    "openrouter": {
      "model": "anthropic/claude-3.5-sonnet"
    }
  }
}
```

**Available Models:**

- `anthropic/claude-3.5-sonnet` (default)
- `openai/gpt-4o`
- `openai/gpt-4o-mini`
- `meta-llama/llama-3.1-405b-instruct`
- `google/gemini-pro-1.5`

#### 2. Groq

Fast inference with various models. Requires a Groq API key.

**Setup:**

```sh
export GROQ_API_KEY="your-groq-api-key"
```

**Configuration:**

```json
{
  "aiprovider": "groq",
  "aichoices": {
    "groq": {
      "model": "mixtral-8x7b-32768"
    }
  }
}
```

**Available Models:**

- `mixtral-8x7b-32768` (default)
- `llama2-70b-4096`
- `llama2-13b-chat`
- `gemma-7b-it`

### Large Diff Handling

Goji automatically handles large diffs using intelligent summarization:

- **Small diffs** (<20k chars): Processed normally
- **Large diffs** (>20k chars): Automatically summarized to key changes only
- **Smart compression**: Reduces large diffs by 90%+ while preserving important context
- **No rate limits**: Single API call prevents token limit issues

### Examples

**Basic commit generation:**

```sh
# Stage your changes
git add .

# Generate commit message
goji draft

# Generate and commit directly
goji draft --commit
```

**Detailed commit with body:**

```sh
# Generate detailed commit with body
goji draft --body --commit
```

**Override type and scope:**

```sh
# Force a specific type and scope
goji draft --type feat --scope api --commit
```

**Using different AI providers:**

```sh
# Switch to Groq (requires GROQ_API_KEY)
goji draft --commit

# Switch to OpenRouter (requires OPENROUTER_API_KEY)  
goji draft --commit
```

### Configuration

Update your `.goji.json` to configure AI providers:

```json
{
  "aiprovider": "openrouter",
  "aichoices": {
    "groq": {
      "model": "mixtral-8x7b-32768"
    },
    "openrouter": {
      "model": "anthropic/claude-3.5-sonnet"
    }
  }
}
```

### Environment Variables

Goji uses environment variables to authenticate with AI providers. Set these in your shell profile (`.bashrc`, `.zshrc`, etc.) or export them before running commands.

#### Required Environment Variables

| Provider | Environment Variable | Required | Description |
|----------|---------------------|----------|-------------|
| **OpenRouter** | `OPENROUTER_API_KEY` | ✅ | Get from [OpenRouter](https://openrouter.ai) |
| **Groq** | `GROQ_API_KEY` | ✅ | Get from [Groq Console](https://console.groq.com) |

**For Groq:**

```sh
export GROQ_API_KEY="your-groq-api-key"
```

**For OpenRouter:**

```sh
export OPENROUTER_API_KEY="your-openrouter-api-key"
```

#### Setting Environment Variables

**Temporary (current session only):**

```sh
export GROQ_API_KEY="your-api-key"
goji draft --commit
```

**Permanent (add to shell profile):**

```sh
# Add to ~/.bashrc, ~/.zshrc, or ~/.profile
echo 'export GROQ_API_KEY="your-api-key"' >> ~/.bashrc
source ~/.bashrc
```

**Using .env files (if supported by your shell):**

```sh
# Create .env file in your project root
echo "GROQ_API_KEY=your-api-key" > .env
```

#### Verifying Environment Variables

Check if your environment variables are set:

```sh
echo $GROQ_API_KEY
echo $OPENROUTER_API_KEY
```

### Troubleshooting

**Rate limiting issues:**

- Goji automatically handles large diffs with smart summarization
- No need to worry about token limits or rate limiting

**API key issues:**

- Ensure your API key is set in the environment
- Check that the provider is correctly configured in `.goji.json`
- Verify environment variables with `echo $VARIABLE_NAME`

**Large diff processing:**

- Goji will show: `🔍 Large diff detected: X chars, using aggressive summarization`
- Summary shows compression: `📊 Summarized to Y chars (Z% reduction)`

**Common environment variable issues:**

- **Variable not found**: Make sure to export the variable in your current shell session
- **Wrong variable name**: Check spelling (GROQ_API_KEY, not GROQ_KEY)
- **Shell restart needed**: Some shells require restarting after adding to profile
- **Case sensitivity**: Environment variable names are case-sensitive

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
