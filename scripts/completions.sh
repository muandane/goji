#!/bin/sh
set -e
rm -rf completions
mkdir completions
for sh in bash zsh fish powershell nushell elvish ion; do
  go run main.go completion "$sh" >"completions/goji.$sh"
done
