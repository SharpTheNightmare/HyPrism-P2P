<p align="right">
  <img src="assets/Hyprism.svg" alt="HyPrism Logo" width="64" height="64" />
</p>

# HyPrism

[![Downloads](https://img.shields.io/github/downloads/yyyumeniku/HyPrism/total?style=for-the-badge&logo=github&label=Downloads&color=207e5c)](https://github.com/yyyumeniku/HyPrism/releases)
[![Website](https://img.shields.io/badge/Website-hyprism-207e5c?style=for-the-badge&logo=firefox)](https://yyyumeniku.github.io/hyprism-site/)
[![Buy Me a Coffee](https://img.shields.io/badge/Buy%20Me%20a%20Coffee-Support-FFDD00?style=for-the-badge&logo=buy-me-a-coffee&logoColor=black)](https://buymeacoffee.com/yyyumeniku)

A multiplatform Hytale launcher with mod manager and more!

<img width="3084" height="1964" alt="Screenshot 2026-01-17 at 22 29 55@2x" src="https://github.com/user-attachments/assets/0a27bc91-d6d5-4148-ae3b-f9e6c36cd6db" />

## Security Notice

**Antivirus false positives are expected for unsigned binaries.** HyPrism is open source and safe. See [SECURITY.md](SECURITY.md) for detailed explanation of why scanners flag it and what HyPrism actually does.

## Installation

### Windows
Download `HyPrism-windows-amd64.exe` from [releases](https://github.com/yyyumeniku/HyPrism/releases)

### macOS
Download `HyPrism.app.zip`, extract and run the app

### Linux

**Recommended:** Use **Flatpak** (bundles all dependencies, no setup required)
```bash
flatpak install HyPrism.flatpak
flatpak run dev.hyprism.HyPrism
```

**Alternative:** AppImage (requires webkit2gtk-4.1)
- Most modern Linux desktops already have WebKit installed
- If not, install `webkit2gtk-4.1` via your package manager before running

**Arch Linux:** Use the AUR package (recommended for Arch users)
```bash
yay -S hyprism-git  # or: paru -S hyprism-git
```
**Note:** DO NOT use the in-app update button with AUR - update only through your AUR helper!

## Platform Support
- ✅ Windows (fully supported)
- ✅ macOS (ARM64)
- ✅ Linux (Flatpak recommended, AppImage for advanced users)
