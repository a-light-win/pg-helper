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

## About compose provider

There are 3 compose providers: `docker-compose`, `podman-compose`, `nerdctl compose`.

But now we need to set secret in the compose file with specific user, group and mode.

| feature                          | docker-compose                             | podman-compose             | nerdctl compose |
| -------------------------------- | ------------------------------------------ | -------------------------- | --------------- |
| secret with user, group and mode | Maybe(using external secret by swarm mode) | Yes(using external secret) | No              |

- docker-compose now report `WARN[0007] secrets `uid`, `gid`and`mode` are not supported, they will be ignored`
  More investigation is needed to see if it is possible to use external secret with docker-compose.
- nerdctl compose do not support external secret yet, and do not support uid, gid, mode fields

because of this, we will use podman-compose as the default compose provider.
