#!/usr/bin/env python3
"""
Web Reconnaissance - Automated web app reconnaissance.
Part of BLACK-HATER cybersecurity skill.

Usage:
    python3 web_recon.py <target-url>
"""

import os
import sys
import subprocess
import json
from pathlib import Path
from urllib.parse import urlparse

class WebRecon:
    """Automated web application reconnaissance."""
    
    def __init__(self, target_url: str):
        self.target = target_url
        self.parsed = urlparse(target_url)
        self.domain = self.parsed.netloc
        self.results = {
            'target': target_url,
            'subdomains': [],
            'directories': [],
            'technologies': [],
            'endpoints': [],
            'headers': {},
            'ssl_info': {},
        }
    
    def run(self):
        """Execute full reconnaissance."""
        print(f"\n{'='*60}")
        print(f"  BLACK-HATER Web Recon")
        print(f"  Target: {self.target}")
        print(f"{'='*60}\n")
        
        self.check_tools()
        self.recon_subdomains()
        self.recon_directories()
        self.recon_technologies()
        self.recon_headers()
        self.recon_endpoints()
        self.generate_report()
    
    def check_tools(self):
        """Check available tools."""
        tools = ['nmap', 'curl', 'ffuf', 'httpx']
        available = []
        
        for tool in tools:
            if self.command_exists(tool):
                available.append(tool)
        
        print(f"[INFO] Available tools: {', '.join(available)}\n")
    
    def command_exists(self, command: str) -> bool:
        """Check if a command exists."""
        try:
            subprocess.run(['which', command], capture_output=True)
            return True
        except FileNotFoundError:
            return False
    
    def recon_subdomains(self):
        """Enumerate subdomains."""
        print("[1/6] Subdomain enumeration...")
        
        # Use httpx for subdomain probing
        if self.command_exists('httpx'):
            try:
                # Simple subdomain wordlist
                subdomains = ['www', 'mail', 'ftp', 'admin', 'api', 'dev', 'staging', 'test']
                found = []
                
                for sub in subdomains:
                    domain = f"{sub}.{self.domain}"
                    result = subprocess.run(
                        ['httpx', '-u', domain, '-sc', '-title', '-silent'],
                        capture_output=True,
                        text=True,
                        timeout=10
                    )
                    if result.stdout.strip():
                        found.append(domain)
                
                self.results['subdomains'] = found
                print(f"  [OK] Found {len(found)} subdomains")
            except Exception as e:
                print(f"  [WARN] Subdomain enum failed: {e}")
        else:
            print("  [SKIP] httpx not available")
    
    def recon_directories(self):
        """Discover directories and files."""
        print("[2/6] Directory discovery...")
        
        # Common paths to check
        common_paths = [
            '/admin', '/login', '/api', '/docs', '/swagger',
            '/.git', '/.env', '/backup', '/config', '/debug',
            '/robots.txt', '/sitemap.xml', '/.well-known'
        ]
        
        found = []
        
        for path in common_paths:
            try:
                result = subprocess.run(
                    ['curl', '-s', '-o', '/dev/null', '-w', '%{http_code}',
                     f'{self.target}{path}'],
                    capture_output=True,
                    text=True,
                    timeout=5
                )
                
                status = result.stdout.strip()
                if status in ['200', '301', '302', '403']:
                    found.append({'path': path, 'status': status})
            except Exception:
                continue
        
        self.results['directories'] = found
        print(f"  [OK] Found {len(found)} paths")
    
    def recon_technologies(self):
        """Detect web technologies."""
        print("[3/6] Technology detection...")
        
        try:
            result = subprocess.run(
                ['curl', '-s', '-I', self.target],
                capture_output=True,
                text=True,
                timeout=10
            )
            
            headers = result.stdout.lower()
            techs = []
            
            # Server detection
            if 'server:' in headers:
                server = [l for l in headers.split('\n') if 'server:' in l]
                if server:
                    techs.append(server[0].split(':', 1)[1].strip())
            
            # Framework detection
            frameworks = {
                'x-powered-by': 'PHP',
                'x-aspnet-version': 'ASP.NET',
                'x-generator': 'Generator',
                'set-cookie': 'Session',
            }
            
            for header, tech in frameworks.items():
                if header in headers:
                    techs.append(tech)
            
            self.results['technologies'] = techs
            print(f"  [OK] Detected: {', '.join(techs) if techs else 'Unknown'}")
            
        except Exception as e:
            print(f"  [WARN] Tech detection failed: {e}")
    
    def recon_headers(self):
        """Analyze HTTP headers."""
        print("[4/6] Header analysis...")
        
        try:
            result = subprocess.run(
                ['curl', '-s', '-I', self.target],
                capture_output=True,
                text=True,
                timeout=10
            )
            
            headers = {}
            for line in result.stdout.split('\n'):
                if ':' in line:
                    key, value = line.split(':', 1)
                    headers[key.strip()] = value.strip()
            
            self.results['headers'] = headers
            
            # Check for security headers
            security_headers = [
                'X-Frame-Options',
                'X-Content-Type-Options',
                'Strict-Transport-Security',
                'Content-Security-Policy',
                'X-XSS-Protection',
            ]
            
            missing = [h for h in security_headers if h not in headers]
            if missing:
                print(f"  [!] Missing security headers: {', '.join(missing)}")
            else:
                print("  [OK] All security headers present")
                
        except Exception as e:
            print(f"  [WARN] Header analysis failed: {e}")
    
    def recon_endpoints(self):
        """Discover API endpoints."""
        print("[5/6] Endpoint discovery...")
        
        # Check common API paths
        api_paths = [
            '/api/v1', '/api/v2', '/graphql', '/rest',
            '/swagger.json', '/openapi.json', '/api-docs',
        ]
        
        found = []
        
        for path in api_paths:
            try:
                result = subprocess.run(
                    ['curl', '-s', '-o', '/dev/null', '-w', '%{http_code}',
                     f'{self.target}{path}'],
                    capture_output=True,
                    text=True,
                    timeout=5
                )
                
                status = result.stdout.strip()
                if status in ['200', '301', '302']:
                    found.append({'path': path, 'status': status})
            except Exception:
                continue
        
        self.results['endpoints'] = found
        print(f"  [OK] Found {len(found)} API endpoints")
    
    def generate_report(self):
        """Generate reconnaissance report."""
        print(f"\n{'='*60}")
        print("  RECONNAISSANCE REPORT")
        print(f"{'='*60}\n")
        
        print(f"Target: {self.target}")
        print(f"Domain: {self.domain}\n")
        
        print("## Subdomains")
        for sub in self.results['subdomains']:
            print(f"  - {sub}")
        if not self.results['subdomains']:
            print("  None found")
        
        print("\n## Discovered Paths")
        for path in self.results['directories']:
            print(f"  [{path['status']}] {path['path']}")
        
        print("\n## Technologies")
        for tech in self.results['technologies']:
            print(f"  - {tech}")
        
        print("\n## API Endpoints")
        for endpoint in self.results['endpoints']:
            print(f"  [{endpoint['status']}] {endpoint['path']}")
        
        print("\n## Security Headers")
        for header in ['X-Frame-Options', 'X-Content-Type-Options', 
                       'Strict-Transport-Security', 'Content-Security-Policy']:
            status = "Present" if header in self.results['headers'] else "MISSING"
            print(f"  {header}: {status}")
        
        # Save report
        report_path = Path("recon_report.json")
        with open(report_path, 'w') as f:
            json.dump(self.results, f, indent=2)
        print(f"\n[OK] Full report saved to {report_path}")


def main():
    if len(sys.argv) < 2:
        print("BLACK-HATER Web Recon")
        print("\nUsage:")
        print("  python3 web_recon.py <target-url>")
        print("\nExamples:")
        print("  python3 web_recon.py https://target.com")
        sys.exit(1)
    
    target = sys.argv[1]
    recon = WebRecon(target)
    recon.run()


if __name__ == '__main__':
    main()
