# Android APK Attack Reference

## Decompilation Techniques

### jadx (Java decompilation)
```bash
jadx -d decompiled target.apk
jadx --deobf target.apk  # Deobfuscate
```

### apktool (Resource extraction)
```bash
apktool d target.apk -o decoded
apktool b decoded -o recompiled.apk  # Recompile
```

### dex2jar (DEX to JAR)
```bash
d2j-dex2jar target.apk -o target-dex2jar.jar
jadx target-dex2jar.jar
```

## Secret Extraction

```bash
# Search for API keys
grep -rn "api_key\|apikey\|API_KEY" decompiled/
grep -rn "secret\|password\|token" decompiled/
grep -rn "BEGIN.*PRIVATE KEY" decompiled/

# Search for URLs
grep -rn "https\?://" decompiled/ | grep -v ".git"

# Search for hardcoded IPs
grep -rn "[0-9]\{1,3\}\.[0-9]\{1,3\}\.[0-9]\{1,3\}\.[0-9]\{1,3\}" decompiled/
```

## Bypass Root Detection

### Frida Script
```javascript
Java.perform(function() {
    var RootDetection = Java.use("com.target.security.RootDetection");
    RootDetection.isRooted.implementation = function() {
        return false;
    };
});
```

### Xposed Module
```java
findAndHookMethod("com.target.security.RootDetection", 
    lpparam.classLoader, 
    "isRooted", 
    new XC_MethodHook() {
        @Override
        protected void afterHookedMethod(MethodHookParam param) {
            param.setResult(false);
        }
    });
```

## Bypass Debugger Detection

### Frida Script
```javascript
Java.perform(function() {
    // Bypass Debug.isDebuggerConnected()
    var Debug = Java.use("android.os.Debug");
    Debug.isDebuggerConnected.implementation = function() {
        return false;
    };
    
    // Bypass ptrace check
    var ptrace = Module.findExportByName(null, "ptrace");
    Interceptor.attach(ptrace, {
        onEnter: function(args) {
            this.original_ptrace = new NativeFunction(ptrace, 'long', ['int', 'int', 'pointer', 'pointer']);
        },
        onLeave: function(retval) {
            retval.replace(0);
        }
    });
});
```

## Bypass SSL Pinning

### Frida Script
```javascript
Java.perform(function() {
    var TrustManager = Java.registerClass({
        name: "com.bypass.TrustManager",
        implements: [Java.use("javax.net.ssl.X509TrustManager")],
        methods: {
            checkClientTrusted: function(chain, authType) {},
            checkServerTrusted: function(chain, authType) {},
            getAcceptedIssuers: function() { return []; }
        }
    });
    
    var SSLContext = Java.use("javax.net.ssl.SSLContext");
    var ctx = SSLContext.getInstance("TLS");
    ctx.init(null, [TrustManager.$new()], null);
});
```

## Extract Encrypted Assets

### TAMK Assets
```bash
python3 scripts/asset_decryptor.py assets/index.html.enc password123
```

### Generic AES Encrypted
```python
from cryptography.hazmat.primitives.ciphers import Cipher, algorithms, modes

def decrypt_aes_gcm(key, iv, ciphertext, tag):
    cipher = Cipher(algorithms.AES(key), modes.GCM(iv, tag))
    decryptor = cipher.decryptor()
    return decryptor.update(ciphertext) + decryptor.finalize()
```

## Repackage APK

```bash
# 1. Decompile
apktool d target.apk -o decoded

# 2. Modify files
# Edit decoded/smali/com/target/...

# 3. Recompile
apktool b decoded -o unsigned.apk

# 4. Sign
keytool -genkey -v -keystore debug.keystore -alias androiddebugkey \
  -keyalg RSA -keysize 2048 -validity 10000
jarsigner -verbose -sigalg SHA1withRSA -digestalg SHA1 \
  -keystore debug.keystore unsigned.apk androiddebugkey

# 5. Align
zipalign -v 4 unsigned.apk target-modified.apk
```

## Hooking with Objection

```bash
# Spawn app
objection -g com.target.app explore

# List activities
android hooking list activities

# Hook method
android hooking watch class com.target.security.GuardianBridge

# Return false to all methods
android hooking watch class com.target.security.GuardianBridge \
  --dump-args --dump-return
```

## Memory Dump

```bash
# Dump app memory
adb shell "cat /proc/$(pidof com.target.app)/mem" > mem.dump

# Search for secrets
strings mem.dump | grep -i "password\|secret\|key"
```

## Network Interception

```bash
# mitmproxy
mitmproxy --mode transparent

# Burp Suite
# Configure proxy: 127.0.0.1:8080
# Install Burp CA certificate

# Intercept traffic
curl -x http://127.0.0.1:8080 -k https://target.com
```
