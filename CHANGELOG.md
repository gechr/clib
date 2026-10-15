# Changelog

## Unreleased

- Add native Cobra completion registration with context-aware predictors, per-command annotation metadata, enum descriptions, boolean values, comma lists, file extensions, and directory filtering. Existing native callbacks and the legacy `--@complete` protocol remain supported.
- Add standalone Bash 3.2+ and native Fish/Zsh clients through the returned completion generator. Fish and Zsh requests preserve literal arguments without evaluation; script patches fail generation if Cobra's templates change.
- Ignore completion action flags inside `__complete` and `__completeNoDesc` requests so completing an uninstall flag cannot remove installed scripts.
