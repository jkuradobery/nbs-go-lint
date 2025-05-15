## GO Linter for [NBS](https://github.com/ydb-platform/nbs) go codestyle

This linter supports the rules from NBS codestyle which are not supported by other linters:
- Indent after `}` (exception: defer is pressed against the block above without indentation)
- Indent before `)` in the case of multiline calls
- Use ID in all identifiers (not Id) (except for protobufs)
