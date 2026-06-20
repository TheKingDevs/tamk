# 📦 Guia de Instalação

Guia de instalação do T.A.M.K em todos os sistemas.

---

## 🐧 Linux

```bash
unzip tamk-proprietary-2026.3.0-HMR-linux.zip
cd tamk
sudo ./install.sh
tamk version
```

👉 [Instruções detalhadas →](INSTALLATION_LINUX.md)

---

## 🪟 Windows

```cmd
REM Extrair e executar install.bat como Admin
tamk version
```

👉 [Instruções detalhadas →](INSTALLATION_WINDOWS.md)

---

## 🍎 macOS

```bash
unzip tamk-proprietary-2026.3.0-HMR-macos.zip
cd tamk
chmod +x install.sh
sudo ./install.sh
tamk version
```

👉 [Instruções detalhadas →](INSTALLATION_MACOS.md)

---

## 📱 Termux (Android)

```bash
pkg install -y golang git openjdk-21 kotlin wget zip apksigner aapt2
git clone https://github.com/Shadw-Developer/tamk.git
cd tamk
go build -o bin/tamk ./cmd/tamk
export PATH="$HOME/tamk/bin:$PATH"
tamk version
```

👉 [Instruções detalhadas →](INSTALLATION_TERMUX.md)

---

## Verificação

```bash
tamk version
tamk --help
tamk create
```

## Desinstalação

```bash
# Linux/macOS
sudo rm -rf /opt/tamk /usr/local/bin/tamk

# Windows
del C:\Windows\System32\tamk.bat
rmdir /s /q "C:\Program Files\TAMK"

# Termux
rm $PREFIX/bin/tamk
rm -rf $PREFIX/opt/tamk ~/.tamk_cache
```

---

## Requisitos

| Recurso | Mínimo | Recomendado |
| :--- | :--- | :--- |
| RAM | 2GB | 8GB |
| Disco | 250MB | 1GB |
| Internet | — | Sim (SDK download) |

---

## Troubleshooting

| Problema | Solução |
| :--- | :--- |
| "comando não encontrado" | Reabra o terminal |
| "permissão negada" | Use `sudo` (Linux/macOS), Admin (Windows) |
| "antivírus bloqueia" | Adicione à lista de exclusão |

---

<div align="center">
  <sub>T.A.M.K v2026.3.0-HMR — Instalação</sub>
</div>
