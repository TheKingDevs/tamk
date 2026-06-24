#!/usr/bin/env python3
"""
APK Security Analyzer - Comprehensive security analysis for Android APKs.
Part of T.A.M.K Android Web Pentester skill.
"""

import os
import sys
import subprocess
import json
import re
from pathlib import Path
from typing import Dict, List, Tuple

class APKAnalyzer:
    """Performs comprehensive security analysis on Android APKs."""
    
    # Sensitive patterns to search for
    SECRET_PATTERNS = [
        r'api[_-]?key\s*[=:]\s*["\']([^"\']+)["\']',
        r'secret[_-]?key\s*[=:]\s*["\']([^"\']+)["\']',
        r'password\s*[=:]\s*["\']([^"\']+)["\']',
        r'token\s*[=:]\s*["\']([^"\']+)["\']',
        r'aws[_-]?access[_-]?key[_-]?id\s*[=:]\s*["\']([^"\']+)["\']',
        r'aws[_-]?secret[_-]?access[_-]?key\s*[=:]\s*["\']([^"\']+)["\']',
        r'private[_-]?key\s*[=:]\s*["\']([^"\']+)["\']',
        r'BEGIN\s+(RSA|DSA|EC)\s+PRIVATE\s+KEY',
    ]
    
    # Insecure configurations
    INSECURE_CONFIGS = [
        (r'allowBackup\s*=\s*"true"', 'Backup enabled - data extraction possible'),
        (r'debuggable\s*=\s*"true"', 'App is debuggable'),
        (r'allowFileAccess\s*=\s*"true"', 'File access enabled'),
        (r'allowUniversalAccessFromFileURLs\s*=\s*"true"', 'Universal file access enabled'),
        (r'cleartextTrafficPermitted\s*=\s*"true"', 'Cleartext traffic allowed'),
        (r'usesCleartextTraffic\s*=\s*"true"', 'Cleartext traffic enabled'),
    ]
    
    # Security implementations to detect
    SECURITY_CHECKS = [
        (r'GuardianBridge', 'Guardian Security System'),
        (r'RASPSecurityModule', 'RASP Protection'),
        (r'isDebuggerAttached', 'Anti-Debug Detection'),
        (r'isDeviceRooted', 'Anti-Root Detection'),
        (r'detectFrida', 'Anti-Frida Detection'),
        (r'isEmulator', 'Anti-Emulator Detection'),
        (r'CertificatePinner', 'Certificate Pinning'),
        (r'NetworkSecurityConfig', 'Network Security Config'),
    ]
    
    def __init__(self, apk_path: str):
        self.apk_path = Path(apk_path)
        self.decompiled_dir = None
        self.findings = {
            'secrets': [],
            'insecure_configs': [],
            'security_features': [],
            'permissions': [],
            'components': [],
            'encrypted_assets': [],
            'urls': [],
        }
    
    def analyze(self) -> Dict:
        """Run full analysis on the APK."""
        print(f"\n{'='*60}")
        print(f"  APK Security Analyzer")
        print(f"  Target: {self.apk_path.name}")
        print(f"{'='*60}\n")
        
        # Step 1: Decompile
        print("[1/6] Decompiling APK...")
        self.decompile()
        
        # Step 2: Analyze manifest
        print("[2/6] Analyzing AndroidManifest.xml...")
        self.analyze_manifest()
        
        # Step 3: Search for secrets
        print("[3/6] Searching for hardcoded secrets...")
        self.search_secrets()
        
        # Step 4: Check configurations
        print("[4/6] Checking insecure configurations...")
        self.check_configurations()
        
        # Step 5: Detect security features
        print("[5/6] Detecting security implementations...")
        self.detect_security_features()
        
        # Step 6: Check encrypted assets
        print("[6/6] Analyzing encrypted assets...")
        self.analyze_encrypted_assets()
        
        # Generate report
        self.generate_report()
        
        return self.findings
    
    def decompile(self):
        """Decompile APK using jadx."""
        if not self.apk_path.exists():
            print(f"  [ERROR] APK not found: {self.apk_path}")
            sys.exit(1)
        
        self.decompiled_dir = self.apk_path.parent / "decompiled"
        
        # Remove old decompilation
        if self.decompiled_dir.exists():
            import shutil
            shutil.rmtree(self.decompiled_dir)
        
        # Run jadx
        try:
            result = subprocess.run(
                ['jadx', '-d', str(self.decompiled_dir), str(self.apk_path)],
                capture_output=True,
                text=True,
                timeout=120
            )
            if result.returncode != 0:
                print(f"  [WARN] jadx returned non-zero: {result.stderr}")
            else:
                print(f"  [OK] Decompiled to {self.decompiled_dir}")
        except FileNotFoundError:
            print("  [ERROR] jadx not found. Install with: pkg install jadx")
            sys.exit(1)
        except subprocess.TimeoutExpired:
            print("  [ERROR] jadx timed out after 120 seconds")
            sys.exit(1)
    
    def analyze_manifest(self):
        """Analyze AndroidManifest.xml for issues."""
        manifest_path = self.decompiled_dir / "resources" / "AndroidManifest.xml"
        if not manifest_path.exists():
            print("  [WARN] AndroidManifest.xml not found")
            return
        
        content = manifest_path.read_text(errors='ignore')
        
        # Extract permissions
        permissions = re.findall(r'android\.permission\.(\w+)', content)
        self.findings['permissions'] = permissions
        
        # Extract components
        components = re.findall(r'android:name="([^"]+)"', content)
        self.findings['components'] = components
        
        # Check for exported components
        exported = re.findall(r'android:exported="true"', content)
        if exported:
            self.findings['insecure_configs'].append({
                'type': 'EXPORTED_COMPONENTS',
                'severity': 'HIGH',
                'description': f'{len(exported)} exported components found'
            })
        
        print(f"  [OK] Found {len(permissions)} permissions, {len(components)} components")
    
    def search_secrets(self):
        """Search for hardcoded secrets in source code."""
        sources_dir = self.decompiled_dir / "sources"
        if not sources_dir.exists():
            return
        
        secrets_found = []
        
        for java_file in sources_dir.rglob("*.java"):
            try:
                content = java_file.read_text(errors='ignore')
                for pattern in self.SECRET_PATTERNS:
                    matches = re.findall(pattern, content, re.IGNORECASE)
                    for match in matches:
                        if len(match) > 3:  # Filter out false positives
                            secrets_found.append({
                                'file': str(java_file.relative_to(sources_dir)),
                                'pattern': pattern,
                                'value': match[:20] + '...' if len(match) > 20 else match
                            })
            except Exception:
                continue
        
        self.findings['secrets'] = secrets_found
        
        if secrets_found:
            print(f"  [!] Found {len(secrets_found)} potential secrets!")
        else:
            print("  [OK] No hardcoded secrets detected")
    
    def check_configurations(self):
        """Check for insecure configurations."""
        issues = []
        
        # Check AndroidManifest.xml
        manifest_path = self.decompiled_dir / "resources" / "AndroidManifest.xml"
        if manifest_path.exists():
            content = manifest_path.read_text(errors='ignore')
            for pattern, description in self.INSECURE_CONFIGS:
                if re.search(pattern, content, re.IGNORECASE):
                    issues.append({
                        'type': 'INSECURE_CONFIG',
                        'severity': 'MEDIUM',
                        'description': description
                    })
        
        # Check network security config
        nsc_path = self.decompiled_dir / "resources" / "res" / "xml" / "network_security_config.xml"
        if nsc_path.exists():
            content = nsc_path.read_text(errors='ignore')
            if 'cleartextTrafficPermitted="true"' in content:
                issues.append({
                    'type': 'CLEARTEXT_TRAFFIC',
                    'severity': 'HIGH',
                    'description': 'Cleartext traffic permitted'
                })
        
        self.findings['insecure_configs'].extend(issues)
        
        if issues:
            print(f"  [!] Found {len(issues)} insecure configurations")
        else:
            print("  [OK] No insecure configurations detected")
    
    def detect_security_features(self):
        """Detect security implementations in code."""
        sources_dir = self.decompiled_dir / "sources"
        if not sources_dir.exists():
            return
        
        features_found = []
        
        for java_file in sources_dir.rglob("*.java"):
            try:
                content = java_file.read_text(errors='ignore')
                for pattern, feature_name in self.SECURITY_CHECKS:
                    if re.search(pattern, content):
                        features_found.append({
                            'feature': feature_name,
                            'file': str(java_file.relative_to(sources_dir))
                        })
            except Exception:
                continue
        
        self.findings['security_features'] = features_found
        
        if features_found:
            print(f"  [OK] Found {len(features_found)} security features")
        else:
            print("  [!] No security features detected")
    
    def analyze_encrypted_assets(self):
        """Analyze encrypted assets in the APK."""
        assets_dir = self.decompiled_dir / "resources" / "assets"
        if not assets_dir.exists():
            return
        
        encrypted_files = []
        
        for asset_file in assets_dir.iterdir():
            if asset_file.suffix == '.enc':
                # Read magic header
                try:
                    with open(asset_file, 'rb') as f:
                        header = f.read(10)
                        if header == b'TAMK_ENC_1':
                            encrypted_files.append({
                                'file': asset_file.name,
                                'size': asset_file.stat().st_size,
                                'encrypted': True,
                                'magic': 'TAMK_ENC_1'
                            })
                        else:
                            encrypted_files.append({
                                'file': asset_file.name,
                                'size': asset_file.stat().st_size,
                                'encrypted': False,
                                'magic': header.decode('ascii', errors='ignore')
                            })
                except Exception:
                    pass
            else:
                # Check if file might be encrypted (no extension or wrong content)
                try:
                    with open(asset_file, 'rb') as f:
                        header = f.read(10)
                        if header == b'TAMK_ENC_1':
                            encrypted_files.append({
                                'file': asset_file.name,
                                'size': asset_file.stat().st_size,
                                'encrypted': True,
                                'magic': 'TAMK_ENC_1'
                            })
                except Exception:
                    pass
        
        self.findings['encrypted_assets'] = encrypted_files
        
        if encrypted_files:
            print(f"  [OK] Found {len(encrypted_files)} encrypted assets")
        else:
            print("  [INFO] No encrypted assets found")
    
    def generate_report(self):
        """Generate analysis report."""
        print(f"\n{'='*60}")
        print("  ANALYSIS REPORT")
        print(f"{'='*60}\n")
        
        # Secrets
        print("## Hardcoded Secrets")
        if self.findings['secrets']:
            for secret in self.findings['secrets'][:5]:  # Show first 5
                print(f"  - {secret['file']}: {secret['value']}")
        else:
            print("  None found")
        
        # Insecure configurations
        print("\n## Insecure Configurations")
        if self.findings['insecure_configs']:
            for config in self.findings['insecure_configs']:
                print(f"  [{config['severity']}] {config['description']}")
        else:
            print("  None found")
        
        # Security features
        print("\n## Security Features Detected")
        if self.findings['security_features']:
            features = set(f['feature'] for f in self.findings['security_features'])
            for feature in features:
                print(f"  - {feature}")
        else:
            print("  None detected")
        
        # Encrypted assets
        print("\n## Encrypted Assets")
        if self.findings['encrypted_assets']:
            for asset in self.findings['encrypted_assets']:
                status = "ENCRYPTED" if asset['encrypted'] else "NOT ENCRYPTED"
                print(f"  - {asset['file']}: {status} ({asset['size']} bytes)")
        else:
            print("  None found")
        
        # Permissions
        print("\n## Permissions")
        if self.findings['permissions']:
            print(f"  Total: {len(self.findings['permissions'])}")
            for perm in self.findings['permissions'][:10]:
                print(f"  - {perm}")
        else:
            print("  None found")
        
        # Save JSON report
        report_path = self.apk_path.parent / "security_report.json"
        with open(report_path, 'w') as f:
            json.dump(self.findings, f, indent=2)
        print(f"\n[OK] Full report saved to {report_path}")


def main():
    if len(sys.argv) < 2:
        print("Usage: python3 apk_analyzer.py <path-to-apk>")
        print("\nExample:")
        print("  python3 apk_analyzer.py app.apk")
        sys.exit(1)
    
    apk_path = sys.argv[1]
    analyzer = APKAnalyzer(apk_path)
    analyzer.analyze()


if __name__ == '__main__':
    main()
