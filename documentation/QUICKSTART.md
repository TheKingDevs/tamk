# 🚀 Quick Start Guide

> This guide takes you from zero to a working WebApp in under 10 minutes.

---

## 📥 Installation

```bash
# 1. Clone the repository
git clone https://github.com/Shadw-Developer/tamk.git
cd tamk

# 2. Build the Go binary
make build
# or: go build -o bin/tamk ./cmd/tamk

# 3. (Optional) Install globally
bash setup-install.sh

# 4. Verify installation
bin/tamk version
```

> 💡 The Go binary is compiled to `bin/tamk`. The `setup-install.sh` script configures dependencies, installs the binary globally, and prepares the environment.

**Manual dependencies (if needed):**
```bash
pkg install -y golang openjdk-21 kotlin wget zip apksigner aapt2 termux-tools git ncurses-utils toilet
```

---

## Step 1: Create a WebApp

```bash
./tamk create
# or use the installed binary: tamk create
```

In the wizard:
1. **Name**: `MyWebApp`
2. **Author**: Your name
3. **Version**: `1.0.0` (default)
4. **Type**: Select `[3] WebApp (WebView + HTML/CSS/JS)`
5. **Content mode**: `[1] Internal (assets/)`
6. **Keystore password**: Create a password (min. 6 chars) and **save it**

---

## Step 2: Add Web Content

```bash
cd MyWebApp
nano src/main/assets/index.html
```

Replace with your HTML or edit the template. Example:

```html
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>My WebApp</title>
    <style>
        body { font-family: Arial; background: #121212; color: white;
               display: flex; align-items: center; justify-content: center;
               height: 100vh; margin: 0; }
        h1 { font-size: 2.5em; }
        button { padding: 15px 30px; background: #6200EE; color: white;
                 border: none; border-radius: 8px; font-size: 18px; cursor: pointer; }
    </style>
</head>
<body>
    <h1>Hello, World!</h1>
    <button onclick="alert('JS working!')">Click</button>
</body>
</html>
```

---

## Step 3: Build APK

```bash
tamk build -p YOUR_PASSWORD
```

The pipeline executes:
1. ✅ Validate keystore password
2. 🔍 Check build cache (SHA-256 hash)
3. 🏗️ Compile resources (aapt2)
4. ☕ Compile Kotlin
5. 📦 Generate DEX (d8)
6. ✍️ Sign + align (apksigner + zipalign)

**Expected output:**
```
✅ SUCCESS: {name}-{version}-release.apk generated successfully!
📦 APK Size: 45.2 KB
```

---

## Step 4: Install and Test

```bash
tamk install
```

Android will open the installer. Confirm installation and open the app.

---

## Step 5: Development with HMR

```bash
# In the project directory
tamk dev
```

Edit files in `src/main/assets/` and see changes in real time.

| File | Behavior |
| :--- | :--- |
| `css/styles.css` | Change logged (HMR-ready, WS server pending) |
| `js/app.js` | Change logged (HMR-ready, WS server pending) |
| `index.html` | Triggers full assets rebuild + APK re-sign + install |

**Dev mode commands:**
| Key | Action |
| :--- | :--- |
| `s` | Status (builds, clients, modules) |
| `b` | Force rebuild |
| `i` | Install APK |
| `q` | Quit |

---

## 🎯 Next Steps

| Action | Description |
| :--- | :--- |
| 🎨 **Use React/Vue** | Compile to static and copy to `assets/` |
| 🖼️ **Customize icon** | Edit `res/drawable/ic_launcher.xml` |
| 📄 **Remote URL** | Create project with external URL mode |
| 🔄 **Update** | `tamk update` to check for new versions |
| ❓ **Help** | Check the [FAQ](FAQ.md) |

---

## 📚 Related Documentation

| Document | Description |
| :--- | :--- |
| [🏗️ ARCHITECTURE.md](ARCHITECTURE.md) | How T.A.M.K works internally |
| [⚙️ DEV_GUIDE.md](DEV_GUIDE.md) | Complete development guide |
| [🔥 HMR_SYSTEM.md](HMR_SYSTEM.md) | Advanced Hot Module Replacement |
| [📋 API_COMPONENTS.md](API_COMPONENTS.md) | API reference |
| [❓ FAQ.md](FAQ.md) | Frequently asked questions |

---

<div align="center">
  <sub>Made with ❤️ by @mrx_dev</sub>
</div>
