![wugo_logo](https://github.com/user-attachments/assets/00770646-61fb-4e1b-98fa-fceeb1cd4aa3)

# wugo — Wallpaper Updater (written on GO)

It's rewritten on Go [wu](https://github.com/kostya1F634/wu) script

## ✨ Features

- 🔄 easy way to update desktop and lock screen wallpaper simultaneously
- 🌐 download wallpapers from URLs or use local images
- ⚙️ automatically organize wallpapers in a dedicated directory
- 🚀 update wallpapers blazingly fast from terminal

## 💡 Idea of Usage

### 🌐 Browsing -> 🖼️ See Image -> 🔄 Update Wallpapers

```sh
wugo https://example.com/image.jpg
wugo image.png
wugo /path/to/image.jpg
wugo file:///path/to/image.jpg
```

## 🧰 Options

Saves/moves the image to custom directory (default ~/wallpapers).

```
wugo -d ~/path/to/dir https://example.com/image.jpg
wugo -d ~/path/to/dir image.png
```

Use local file without moving it to wallpapers directory.

```
wugo -nm image.png
```

## 🔧 Installation from Source

### 📋 Requirements

- 🛠️ make
- 🦫 Go

```
git clone https://github.com/kostya1F634/wugo.git
cd wugo
make bin
# binary in bin directory
```

## 🧪 Development & Testing

```sh
go run ./cmd/wugo <image>
go test ./...
```
