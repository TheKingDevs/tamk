#!/usr/bin/env python3
"""
Vulnerability Scanner - OWASP Top 10 scanner for Android/Web apps.
Part of T.A.M.K Android Web Pentester skill.

Usage:
    python3 vuln_scanner.py <target-directory>
    python3 vuln_scanner.py --url <url>
"""

import os
import sys
import re
import json
from pathlib import Path
from typing import Dict, List
from urllib.parse import urlparse

# OWASP Top 10 (2021)
OWASP_TOP_10 = {
    'A01': 'Broken Access Control',
    'A02': 'Cryptographic Failures',
    'A03': 'Injection',
    'A04': 'Insecure Design',
    'A05': 'Security Misconfiguration',
    'A06': 'Vulnerable and Outdated Components',
    'A07': 'Identification and Authentication Failures',
    'A08': 'Software and Data Integrity Failures',
    'A09': 'Security Logging and Monitoring Failures',
    'A10': 'Server-Side Request Forgery (SSRF)',
}


class VulnerabilityScanner:
    """Scans for OWASP Top 10 vulnerabilities."""
    
    def __init__(self, target: str):
        self.target = Path(target) if not target.startswith('http') else target
        self.vulnerabilities = []
    
    def scan(self) -> List[Dict]:
        """Run full vulnerability scan."""
        print(f"\n{'='*60}")
        print("  Vulnerability Scanner (OWASP Top 10)")
        print(f"  Target: {self.target}")
        print(f"{'='*60}\n")
        
        if isinstance(self.target, Path):
            self.scan_directory()
        else:
            self.scan_url()
        
        self.generate_report()
        return self.vulnerabilities
    
    def scan_directory(self):
        """Scan a directory for vulnerabilities."""
        if not self.target.exists():
            print(f"[ERROR] Directory not found: {self.target}")
            return
        
        # Find source files
        source_files = list(self.target.rglob("*.java")) + \
                       list(self.target.rglob("*.kt")) + \
                       list(self.target.rglob("*.js")) + \
                       list(self.target.rglob("*.html"))
        
        print(f"[INFO] Scanning {len(source_files)} files...\n")
        
        for source_file in source_files:
            try:
                content = source_file.read_text(errors='ignore')
                self.scan_content(content, str(source_file))
            except Exception:
                continue
    
    def scan_url(self):
        """Scan a URL for vulnerabilities."""
        print("[INFO] URL scanning not yet implemented")
        print("[INFO] Use directory scanning for decompiled APKs")
    
    def scan_content(self, content: str, file_path: str):
        """Scan file content for vulnerabilities."""
        # A01: Broken Access Control
        self.check_access_control(content, file_path)
        
        # A02: Cryptographic Failures
        self.check_crypto_failures(content, file_path)
        
        # A03: Injection
        self.check_injection(content, file_path)
        
        # A05: Security Misconfiguration
        self.check_misconfiguration(content, file_path)
        
        # A07: Authentication Failures
        self.check_auth_failures(content, file_path)
    
    def check_access_control(self, content: str, file_path: str):
        """Check for broken access control vulnerabilities."""
        issues = []
        
        # Check for hardcoded credentials
        if re.search(r'password\s*=\s*["\'][^"\']+["\']', content, re.IGNORECASE):
            issues.append({
                'owasp': 'A01',
                'severity': 'HIGH',
                'title': 'Hardcoded Password',
                'file': file_path,
                'description': 'Password hardcoded in source code'
            })
        
        # Check for exposed admin paths
        if re.search(r'/admin|/manage|/console', content, re.IGNORECASE):
            issues.append({
                'owasp': 'A01',
                'severity': 'MEDIUM',
                'title': 'Admin Path Exposed',
                'file': file_path,
                'description': 'Admin path found in source code'
            })
        
        self.vulnerabilities.extend(issues)
    
    def check_crypto_failures(self, content: str, file_path: str):
        """Check for cryptographic failures."""
        issues = []
        
        # Check for weak hashing algorithms
        if re.search(r'MD5|SHA1', content):
            issues.append({
                'owasp': 'A02',
                'severity': 'MEDIUM',
                'title': 'Weak Hashing Algorithm',
                'file': file_path,
                'description': 'MD5 or SHA1 detected - use SHA-256 or better'
            })
        
        # Check for hardcoded encryption keys
        if re.search(r'key\s*=\s*["\'][A-Za-z0-9+/=]{16,}["\']', content):
            issues.append({
                'owasp': 'A02',
                'severity': 'HIGH',
                'title': 'Hardcoded Encryption Key',
                'file': file_path,
                'description': 'Encryption key hardcoded in source'
            })
        
        # Check for ECB mode
        if re.search(r'AES/ECB|DES/ECB', content):
            issues.append({
                'owasp': 'A02',
                'severity': 'HIGH',
                'title': 'Weak Cipher Mode',
                'file': file_path,
                'description': 'ECB mode detected - use CBC or GCM'
            })
        
        self.vulnerabilities.extend(issues)
    
    def check_injection(self, content: str, file_path: str):
        """Check for injection vulnerabilities."""
        issues = []
        
        # SQL Injection
        if re.search(r'executeQuery.*\+|Statement.*execute', content):
            issues.append({
                'owasp': 'A03',
                'severity': 'CRITICAL',
                'title': 'Potential SQL Injection',
                'file': file_path,
                'description': 'String concatenation in SQL query'
            })
        
        # Command Injection
        if re.search(r'Runtime\.getRuntime\(\)\.exec\(|ProcessBuilder.*\+', content):
            issues.append({
                'owasp': 'A03',
                'severity': 'CRITICAL',
                'title': 'Potential Command Injection',
                'file': file_path,
                'description': 'User input may be passed to command execution'
            })
        
        # XSS (in web contexts)
        if re.search(r'innerHTML|document\.write|eval\(', content):
            issues.append({
                'owasp': 'A03',
                'severity': 'HIGH',
                'title': 'Potential XSS Vulnerability',
                'file': file_path,
                'description': 'Unescaped output detected'
            })
        
        self.vulnerabilities.extend(issues)
    
    def check_misconfiguration(self, content: str, file_path: str):
        """Check for security misconfigurations."""
        issues = []
        
        # Debug mode in production
        if re.search(r'debug\s*=\s*true|DEBUG\s*=\s*true', content):
            issues.append({
                'owasp': 'A05',
                'severity': 'MEDIUM',
                'title': 'Debug Mode Enabled',
                'file': file_path,
                'description': 'Debug mode may be enabled in production'
            })
        
        # Verbose error messages
        if re.search(r'printStackTrace\(\)|e\.getMessage\(\)', content):
            issues.append({
                'owasp': 'A05',
                'severity': 'LOW',
                'title': 'Verbose Error Messages',
                'file': file_path,
                'description': 'Stack traces may leak sensitive information'
            })
        
        self.vulnerabilities.extend(issues)
    
    def check_auth_failures(self, content: str, file_path: str):
        """Check for authentication failures."""
        issues = []
        
        # Weak password validation
        if re.search(r'password\.length\s*<\s*[0-9]', content):
            issues.append({
                'owasp': 'A07',
                'severity': 'MEDIUM',
                'title': 'Weak Password Policy',
                'file': file_path,
                'description': 'Password length requirement may be too short'
            })
        
        # Missing authentication
        if re.search(r'android:exported="true"', content):
            issues.append({
                'owasp': 'A07',
                'severity': 'HIGH',
                'title': 'Exported Component Without Auth',
                'file': file_path,
                'description': 'Component is exported without authentication'
            })
        
        self.vulnerabilities.extend(issues)
    
    def generate_report(self):
        """Generate vulnerability report."""
        print(f"\n{'='*60}")
        print("  VULNERABILITY REPORT")
        print(f"{'='*60}\n")
        
        # Group by OWASP category
        by_category = {}
        for vuln in self.vulnerabilities:
            category = vuln['owasp']
            if category not in by_category:
                by_category[category] = []
            by_category[category].append(vuln)
        
        # Print by category
        for category, vulns in sorted(by_category.items()):
            print(f"## {category}: {OWASP_TOP_10.get(category, 'Unknown')}")
            for vuln in vulns:
                severity_icon = {
                    'CRITICAL': '🔴',
                    'HIGH': '🟠',
                    'MEDIUM': '🟡',
                    'LOW': '🟢'
                }.get(vuln['severity'], '⚪')
                
                print(f"  {severity_icon} [{vuln['severity']}] {vuln['title']}")
                print(f"     {vuln['description']}")
                print(f"     File: {vuln['file']}")
            print()
        
        # Summary
        print("## Summary")
        print(f"  Total vulnerabilities: {len(self.vulnerabilities)}")
        
        by_severity = {}
        for vuln in self.vulnerabilities:
            severity = vuln['severity']
            by_severity[severity] = by_severity.get(severity, 0) + 1
        
        for severity in ['CRITICAL', 'HIGH', 'MEDIUM', 'LOW']:
            if severity in by_severity:
                print(f"  {severity}: {by_severity[severity]}")
        
        # Save JSON report
        report_path = Path("vulnerability_report.json")
        with open(report_path, 'w') as f:
            json.dump({
                'target': str(self.target),
                'vulnerabilities': self.vulnerabilities,
                'summary': {
                    'total': len(self.vulnerabilities),
                    'by_severity': by_severity,
                    'by_category': {k: len(v) for k, v in by_category.items()}
                }
            }, f, indent=2)
        
        print(f"\n[OK] Full report saved to {report_path}")


def main():
    if len(sys.argv) < 2:
        print("Vulnerability Scanner (OWASP Top 10)")
        print("\nUsage:")
        print("  python3 vuln_scanner.py <target-directory>")
        print("\nExamples:")
        print("  python3 vuln_scanner.py ./decompiled/sources/")
        sys.exit(1)
    
    target = sys.argv[1]
    scanner = VulnerabilityScanner(target)
    scanner.scan()


if __name__ == '__main__':
    main()
