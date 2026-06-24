#!/usr/bin/env python3
"""
Secret Extractor - Extract hardcoded secrets from any codebase.
Part of BLACK-HATER cybersecurity skill.

Usage:
    python3 secret_extractor.py <directory>
    python3 secret_extractor.py --file <file>
"""

import os
import sys
import re
import json
from pathlib import Path
from typing import Dict, List

class SecretExtractor:
    """Extracts hardcoded secrets from source code."""
    
    # Comprehensive secret patterns
    PATTERNS = {
        'API Key': [
            r'api[_-]?key\s*[=:]\s*["\']([^"\']{8,})["\']',
            r'apikey\s*[=:]\s*["\']([^"\']{8,})["\']',
            r'API_KEY\s*[=:]\s*["\']([^"\']{8,})["\']',
            r'["\']([A-Za-z0-9]{32,})["\']',  # Generic long string
        ],
        'AWS Key': [
            r'aws[_-]?access[_-]?key[_-]?id\s*[=:]\s*["\']?(AKIA[A-Z0-9]{16})["\']?',
            r'aws[_-]?secret[_-]?access[_-]?key\s*[=:]\s*["\']([^"\']{40})["\']',
        ],
        'Private Key': [
            r'BEGIN\s+(RSA|DSA|EC|OPENSSH)\s+PRIVATE\s+KEY',
            r'private[_-]?key\s*[=:]\s*["\']([^"\']{20,})["\']',
        ],
        'Password': [
            r'password\s*[=:]\s*["\']([^"\']{3,})["\']',
            r'passwd\s*[=:]\s*["\']([^"\']{3,})["\']',
            r'pwd\s*[=:]\s*["\']([^"\']{3,})["\']',
            r'secret[_-]?password\s*[=:]\s*["\']([^"\']{3,})["\']',
        ],
        'Token': [
            r'token\s*[=:]\s*["\']([^"\']{8,})["\']',
            r'auth[_-]?token\s*[=:]\s*["\']([^"\']{8,})["\']',
            r'bearer\s+[A-Za-z0-9\-._~+/]+=*',  # JWT-like
            r'eyJ[A-Za-z0-9\-._~+/]+=*',  # JWT
        ],
        'Database': [
            r'jdbc:[a-z]+://[^\s"\']+',
            r'mongodb(\+srv)?://[^\s"\']+',
            r'postgres(ql)?://[^\s"\']+',
            r'mysql://[^\s"\']+',
            r'redis://[^\s"\']+',
        ],
        'Cloud': [
            r'sk_live_[A-Za-z0-9]+',  # Stripe
            r'pk_live_[A-Za-z0-9]+',  # Stripe
            r'sg_[A-Za-z0-9]+',  # SendGrid
            r'xox[baprs]-[A-Za-z0-9]+',  # Slack
            r'ghp_[A-Za-z0-9]+',  # GitHub
            r'glpat-[A-Za-z0-9\-]+',  # GitLab
        ],
        'IP Address': [
            r'\b(?:[0-9]{1,3}\.){3}[0-9]{1,3}\b',
        ],
        'URL with credentials': [
            r'https?://[^:]+:[^@]+@[^\s"\']+',
        ],
        'Hash': [
            r'[a-fA-F0-9]{32,}',  # MD5/SHA
            r'\$2[aby]?\$\d{1,2}\$.{53}',  # bcrypt
        ],
    }
    
    def __init__(self):
        self.findings = []
    
    def extract_from_directory(self, directory: str) -> List[Dict]:
        """Extract secrets from all files in a directory."""
        directory = Path(directory)
        
        if not directory.exists():
            print(f"[ERROR] Directory not found: {directory}")
            return []
        
        # File extensions to scan
        extensions = {'.java', '.kt', '.js', '.ts', '.py', '.go', '.rb', 
                     '.php', '.xml', '.json', '.yaml', '.yml', '.env', '.config'}
        
        files = []
        for ext in extensions:
            files.extend(directory.rglob(f"*{ext}"))
        
        print(f"\n[INFO] Scanning {len(files)} files...\n")
        
        for file_path in files:
            self.scan_file(file_path)
        
        return self.findings
    
    def extract_from_file(self, file_path: str) -> List[Dict]:
        """Extract secrets from a single file."""
        return self.scan_file(Path(file_path))
    
    def scan_file(self, file_path: Path) -> List[Dict]:
        """Scan a file for secrets."""
        try:
            content = file_path.read_text(errors='ignore')
        except Exception:
            return []
        
        file_findings = []
        
        for secret_type, patterns in self.PATTERNS.items():
            for pattern in patterns:
                matches = re.finditer(pattern, content, re.IGNORECASE)
                for match in matches:
                    # Get the matched value
                    value = match.group(1) if match.lastindex else match.group(0)
                    
                    # Skip common false positives
                    if self.is_false_positive(value, secret_type):
                        continue
                    
                    # Get line number
                    line_num = content[:match.start()].count('\n') + 1
                    
                    finding = {
                        'type': secret_type,
                        'value': value[:50] + '...' if len(value) > 50 else value,
                        'file': str(file_path),
                        'line': line_num,
                        'context': content[max(0, match.start()-20):match.end()+20].strip()
                    }
                    
                    file_findings.append(finding)
                    self.findings.append(finding)
        
        if file_findings:
            print(f"  [!] {file_path.name}: {len(file_findings)} secrets found")
        
        return file_findings
    
    def is_false_positive(self, value: str, secret_type: str) -> bool:
        """Check if a match is likely a false positive."""
        # Common false positives
        false_positives = [
            'example', 'test', 'dummy', 'placeholder', 'xxx',
            'your_key_here', 'insert_key', 'TODO', 'FIXME',
            '1234567890', 'abcdefghij', 'password123',
            'null', 'undefined', 'none', 'empty',
        ]
        
        value_lower = value.lower()
        
        for fp in false_positives:
            if fp in value_lower:
                return True
        
        # Skip very short values
        if len(value) < 4:
            return True
        
        # Skip common patterns
        if re.match(r'^[a-z_]+$', value):  # Only lowercase letters and underscore
            return True
        
        return False
    
    def generate_report(self) -> str:
        """Generate a report of all findings."""
        report = []
        report.append("\n" + "="*60)
        report.append("  SECRET EXTRACTION REPORT")
        report.append("="*60 + "\n")
        
        # Group by type
        by_type = {}
        for finding in self.findings:
            secret_type = finding['type']
            if secret_type not in by_type:
                by_type[secret_type] = []
            by_type[secret_type].append(finding)
        
        # Print by type
        for secret_type, findings in sorted(by_type.items()):
            report.append(f"## {secret_type}")
            for finding in findings:
                report.append(f"  File: {finding['file']}")
                report.append(f"  Line: {finding['line']}")
                report.append(f"  Value: {finding['value']}")
                report.append(f"  Context: {finding['context'][:80]}...")
                report.append("")
        
        # Summary
        report.append("## Summary")
        report.append(f"  Total secrets found: {len(self.findings)}")
        for secret_type, findings in sorted(by_type.items()):
            report.append(f"  {secret_type}: {len(findings)}")
        
        return "\n".join(report)


def main():
    if len(sys.argv) < 2:
        print("Secret Extractor")
        print("\nUsage:")
        print("  python3 secret_extractor.py <directory>")
        print("  python3 secret_extractor.py --file <file>")
        print("\nExamples:")
        print("  python3 secret_extractor.py ./decompiled/sources/")
        print("  python3 secret_extractor.py --file config.json")
        sys.exit(1)
    
    extractor = SecretExtractor()
    
    if sys.argv[1] == '--file':
        if len(sys.argv) < 3:
            print("[ERROR] Usage: --file <file>")
            sys.exit(1)
        findings = extractor.extract_from_file(sys.argv[2])
    else:
        findings = extractor.extract_from_directory(sys.argv[1])
    
    # Print report
    print(extractor.generate_report())
    
    # Save JSON
    report_path = Path("secrets_report.json")
    with open(report_path, 'w') as f:
        json.dump(extractor.findings, f, indent=2)
    print(f"\n[OK] Full report saved to {report_path}")


if __name__ == '__main__':
    main()
