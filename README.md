![wugo_logo](https://github.com/user-attachments/assets/00770646-61fb-4e1b-98fa-fceeb1cd4aa3)

# wugo — Wallpaper Updater (written on GO)
It's rewrited on Go [wu](https://github.com/kostya1F634/wu) script
## ✨ Features
* 🔄 easy way to update desktop and lock screen wallpaper simultaneously
* ⚙️ automatically download updated wallpaper to directory with all wallpapers
* 🚀 update wallpapers blazingly fast from terminal
## 💡 Idea of Usage
### 🌐 Browsing -> 🖼️ See Image -> 🔄 Update Wallpapers
```sh
wugo https://image-url.ext](https://example.com/image.jpg)
```
## 🧰 Options
Saves the image to custom directory (default ~/wallpapers).
```sh
wugo -d ~/path/to/dir https://example.com/image.jpg
```
Temporarily saves the image to /tmp, without keeping it permanently.
```sh
wugo -ns https://example.com/image.jpg
```
## 🔧 Installation from Source
### 📋 Requirements
* 🛠️ make
* 🦫 Go
```sh
git clone https://github.com/kostya1F634/wugo.git
cd wugo
make bin
# binary in bin direcory
```
