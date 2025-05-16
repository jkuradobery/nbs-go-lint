# GO Linter for [NBS](https://github.com/ydb-platform/nbs) go codestyle

## This linter supports the rules from NBS codestyle which are not supported by other linters:

### Formating
- New line after `}` (exception: defer is pressed against the block above without indentation).
- New line before `)` in the case of multiline calls.
- New line after multiline function signature.
- Use ID in all identifiers (not Id) (except for protobufs).
- If an expression can fit on one line, it should be on one line.

### Separators
- The separator `/////` 80 symbols length is required after package declaration.
- The separator is required between private and public methods.
- The separator is required before and after interface declaration.
- The separator is required around each struct + its methods.
- The separator is forbidden at the end of the file.
- Function groups should also be separated from class methods.
- Use of separators between struct methods is allowed but not mandatory, separators shall be used for separating "logical" groups of methods.
