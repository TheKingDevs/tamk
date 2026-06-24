# iOS Application Attack Reference

## Decompilation

### class-dump
```bash
class-dump Target.app > classes.h
```

### Hopper / IDA Pro
```bash
# Load binary for analysis
hopper Target.app/Target
```

### FlexDecrypt
```bash
# Decrypt iOS app
flexdecrypt Target.app
```

## Jailbreak Detection Bypass

### Frida Script
```bash
# Bypass jailbreak detection
frida -U -f com.target.app -l jailbreak_bypass.js
```

```javascript
// jailbreak_bypass.js
Interceptor.attach(Module.findExportByName('libSystem.B.dylib', 'stat'), {
    onEnter: function(args) {
        this.path = args[0];
    },
    onLeave: function(retval) {
        if (this.path.readUtf8String().indexOf('/Applications/Cydia.app') !== -1 ||
            this.path.readUtf8String().indexOf('/Library/MobileSubstrate') !== -1 ||
            this.path.readUtf8String().indexOf('/bin/bash') !== -1) {
            retval.replace(-1);
        }
    }
});
```

### Objection
```bash
# Spawn and bypass
objection -g com.target.app explore

# Disable jailbreak checks
ios jailbreak disable
```

## SSL Pinning Bypass

### Frida
```javascript
// ssl_bypass.js
var SSLContext = ObjC.classes.NSURLSession;

Interceptor.attach(ObjC.classes.NSURLSession['- dataTaskWithRequest:completionHandler:'].implementation, {
    onEnter: function(args) {
        // Disable SSL pinning
    }
});
```

### Objection
```bash
ios sslpinning disable
```

## Keychain Access

### Keychain Dumper
```bash
# Dump keychain
keychain_dump
```

### Frida
```javascript
// keychain_dump.js
var SecItemCopyMatching = ObjC.classes.Security['SecItemCopyMatching'];
// Hook and extract keychain items
```

## URL Scheme Attacks

```bash
# Test custom URL schemes
open target://callback?token=stolen

# Intercept URL schemes
# Use Burp or custom handler
```

## App Transport Security

```bash
# Check ATS configuration
plutil -p Target.app/Info.plist | grep NSAppTransportSecurity

# Bypass ATS (if configured)
# Add exception domains
```

## Binary Analysis

```bash
# strings extraction
strings Target.app/Target | grep -i "password\|secret\|key"

# otool analysis
otool -L Target.app/Target  # List libraries
otool -tV Target.app/Target  # Disassemble

# nm for symbols
nm Target.app/Target | grep -i "password\|secret"
```

## Runtime Manipulation

### Frida
```javascript
// Hook Objective-C method
var targetClass = ObjC.classes.TargetClass;
Interceptor.attach(targetClass['- verifyToken'].implementation, {
    onLeave: function(retval) {
        retval.replace(0x1);  // Return true
    }
});
```

### Cycript
```bash
# Attach to process
cycript -p Target

# Override method
cy# [[TargetClass alloc] init]
```

## Data Protection

```bash
# Check file protection
ls -la Target.app/Documents/
xattr -l Target.app/Documents/file.db

# Bypass data protection
# Requires jailbreak
```
