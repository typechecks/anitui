anitui
=====

a tui for browsing and streaming anime. scrapes from multiple sources,
navigate with vim keys or arrows, and play in your preferred video player.

quick start
-----------

```bash
# build
make build

# run
./build/anitui

# or via Go
go run ./cmd/anitui
```

controls
--------

| key | action |
|-----|--------|
| enter | confirm / select |
| esc | return |
| j / ↓ | down |
| k / ↑ | up |
| g g | jump to top |
| G | jump to bottom |
| ctrl+u | page up |
| ctrl+d | page down |
| / | search |
| ctrl+c | exit |

player support
--------------

anitui auto-detects and uses:

1. mpv
2. vlc
3. haruna

override via `ANITUI_PLAYER` environment variable.

building from source
--------------------

```bash
git clone https://github.com/typechecks/anitui.git
cd anitui
make build
```

cross-compile:

```bash
make build-linux    # linux amd64 binary
make build-windows  # windows amd64 binary
```
