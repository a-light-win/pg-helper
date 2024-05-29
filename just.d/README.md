# just files

## common commands

- build: `just build`
- unit test: `just test`
- end to end test `just e2e`
- create a migration: `just migration <name>`
- list all commands: `just -l`

## mods directory

The `mods/` directory contains the just modules.

The naming of recipes in the modules should start with alpha character and separate words by the `-` character.
This will avoid any conflicts with the recipes in the library.

## lib directory

The lib/ directory contains the just library. All logic should be written here
so that we can use it as a dependency in other recipes. As of written the just can
not use recipes in modules as dependencies.

The naming of recipes in the library should start with `_` and separate words by the `_` character.
