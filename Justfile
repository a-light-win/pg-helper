
import "just.d/lib/vars.just"

# mod db "just.d/mods/db.just"
mod e2e "just.d/mods/e2e.just"
mod setup "just.d/mods/setup.just"
mod dev "just.d/mods/dev.just"
mod build "just.d/mods/build.just"
mod tag "just.d/mods/tag.just"

list:
  just --unstable -l --list-submodules
