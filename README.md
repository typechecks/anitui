anitui
=====

a tui for browsing and streaming anime. scrapes from multiple sources,
navigate with vim keys or arrows, and play in your preferred video player.

installation
------------

### linux / macos

```bash
curl -sS https://raw.githubusercontent.com/typechecks/anitui/main/scripts/install.sh | sudo sh
```

### windows (powershell)

```powershell
iwr https://raw.githubusercontent.com/typechecks/anitui/main/scripts/install.ps1 -useb | iex
```

### package managers

- **aur**: `yay -S anitui-bin` (prebuilt) or `yay -S anitui` (build from source)
- **winget**: `winget install typechecks.anitui`

platform support
----------------

anitui is built for **linux**, **macos** (intel & apple silicon), and **windows**.

| os | architectures |
|----|---------------|
| linux | amd64, arm64 |
| macos | amd64, arm64 |
| windows | amd64, arm64 |

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

### navigation

| key | action |
|-----|--------|
| enter | confirm / select |
| esc | return |
| j / ↓ | down |
| k / ↑ | up |
| g g | jump to top |
| G | jump to bottom |
| ctrl+u / ⌘u | page up |
| ctrl+d / ⌘d | page down |
| / | search |
| ? | toggle help popup |
| ctrl+c / ⌘c | exit |

### episode screen

| key | action |
|-----|--------|
| j / ↓ | down |
| k / ↑ | up |
| enter | play episode |
| space | toggle synopsis expand |
| d | toggle sub / dub |
| esc | back to results |

### watching screen

| key | action |
|-----|--------|
| h / ← | previous episode |
| l / → | next episode |
| r | replay current episode |
| space | replay current episode |
| s | cycle video source |
| d | toggle sub / dub |
| esc | back to episode list |

player support
--------------

anitui auto-detects and uses:

1. mpv
2. iina (macOS)
3. vlc
4. haruna

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
make build-linux-amd64    # linux x86_64
make build-linux-arm64    # linux arm64
make build-windows-amd64  # windows x86_64
make build-all            # build for all supported platforms
```
