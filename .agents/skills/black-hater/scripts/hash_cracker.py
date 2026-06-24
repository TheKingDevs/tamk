#!/usr/bin/env python3
"""
Hash Cracker - Multi-algorithm hash cracking utility.
Part of BLACK-HATER cybersecurity skill.

Usage:
    python3 hash_cracker.py <hash>
    python3 hash_cracker.py --file <hashes.txt>
    python3 hash_cracker.py --type md5 --wordlist rockyou.txt
"""

import os
import sys
import hashlib
import itertools
import string
from pathlib import Path
from typing import Optional, List

# Try to import for faster cracking
try:
    import multiprocessing as mp
    HAS_MP = True
except ImportError:
    HAS_MP = False


class HashCracker:
    """Multi-algorithm hash cracker."""
    
    ALGORITHMS = {
        'md5': hashlib.md5,
        'sha1': hashlib.sha1,
        'sha256': hashlib.sha256,
        'sha512': hashlib.sha512,
    }
    
    def __init__(self):
        self.cracked = []
    
    def identify_hash(self, hash_str: str) -> List[str]:
        """Identify hash type by length and pattern."""
        hash_str = hash_str.strip().lower()
        length = len(hash_str)
        
        possible = []
        
        # MD5: 32 hex chars
        if length == 32 and all(c in string.hexdigits for c in hash_str):
            possible.append('md5')
        
        # SHA1: 40 hex chars
        if length == 40 and all(c in string.hexdigits for c in hash_str):
            possible.append('sha1')
        
        # SHA256: 64 hex chars
        if length == 64 and all(c in string.hexdigits for c in hash_str):
            possible.append('sha256')
        
        # SHA512: 128 hex chars
        if length == 128 and all(c in string.hexdigits for c in hash_str):
            possible.append('sha512')
        
        # bcrypt: starts with $2a$, $2b$, $2y$
        if hash_str.startswith(('$2a$', '$2b$', '$2y$')):
            possible.append('bcrypt')
        
        # NTLM: 32 hex chars (same as MD5 but different context)
        if length == 32:
            possible.append('ntlm')
        
        return possible if possible else ['unknown']
    
    def crack_wordlist(self, hash_str: str, algorithm: str, wordlist_path: str) -> Optional[str]:
        """Crack hash using wordlist."""
        if not Path(wordlist_path).exists():
            print(f"[ERROR] Wordlist not found: {wordlist_path}")
            return None
        
        hash_func = self.ALGORITHMS.get(algorithm)
        if not hash_func:
            print(f"[ERROR] Unknown algorithm: {algorithm}")
            return None
        
        hash_bytes = bytes.fromhex(hash_str) if algorithm != 'bcrypt' else hash_str
        
        print(f"[*] Cracking {algorithm} hash with wordlist...")
        
        with open(wordlist_path, 'r', errors='ignore') as f:
            for i, line in enumerate(f):
                word = line.strip()
                
                if not word:
                    continue
                
                # Try different encodings
                for encoding in ['utf-8', 'latin-1', 'ascii']:
                    try:
                        word_bytes = word.encode(encoding)
                        computed = hash_func(word_bytes).hexdigest()
                        
                        if computed == hash_str.lower():
                            print(f"[+] FOUND: {word}")
                            return word
                    except:
                        continue
                
                # Progress indicator
                if i % 100000 == 0:
                    print(f"[*] Tried {i} passwords...", end='\r')
        
        print("[-] Not found in wordlist")
        return None
    
    def crack_bruteforce(self, hash_str: str, algorithm: str, 
                         max_length: int = 6, charset: str = None) -> Optional[str]:
        """Crack hash using brute force."""
        if charset is None:
            charset = string.ascii_lowercase + string.digits
        
        hash_func = self.ALGORITHMS.get(algorithm)
        if not hash_func:
            print(f"[ERROR] Unknown algorithm: {algorithm}")
            return None
        
        print(f"[*] Brute forcing {algorithm} hash (max length: {max_length})...")
        
        for length in range(1, max_length + 1):
            print(f"[*] Trying length {length}...")
            
            for attempt in itertools.product(charset, repeat=length):
                word = ''.join(attempt)
                computed = hash_func(word.encode('utf-8')).hexdigest()
                
                if computed == hash_str.lower():
                    print(f"[+] FOUND: {word}")
                    return word
        
        print("[-] Not found via brute force")
        return None
    
    def crack_mask(self, hash_str: str, algorithm: str, mask: str) -> Optional[str]:
        """Crack hash using mask attack.
        
        Mask examples:
            ?d = digit
            ?l = lowercase
            ?u = uppercase
            ?s = special
            ?a = all
        """
        hash_func = self.ALGORITHMS.get(algorithm)
        if not hash_func:
            print(f"[ERROR] Unknown algorithm: {algorithm}")
            return None
        
        charsets = {
            '?d': string.digits,
            '?l': string.ascii_lowercase,
            '?u': string.ascii_uppercase,
            '?s': string.punctuation,
            '?a': string.ascii_letters + string.digits + string.punctuation,
        }
        
        # Parse mask into character sets
        mask_parts = []
        i = 0
        while i < len(mask):
            if mask[i] == '?' and i + 1 < len(mask):
                mask_parts.append(charsets.get(mask[i+1], ''))
                i += 2
            else:
                mask_parts.append(mask[i])
                i += 1
        
        print(f"[*] Mask attack: {mask}")
        
        for attempt in itertools.product(*mask_parts):
            word = ''.join(attempt)
            computed = hash_func(word.encode('utf-8')).hexdigest()
            
            if computed == hash_str.lower():
                print(f"[+] FOUND: {word}")
                return word
        
        print("[-] Not found via mask attack")
        return None
    
    def generate_report(self):
        """Generate cracking report."""
        print(f"\n{'='*60}")
        print("  HASH CRACKING REPORT")
        print(f"{'='*60}\n")
        
        if self.cracked:
            print("## Cracked Hashes")
            for item in self.cracked:
                print(f"  Hash: {item['hash']}")
                print(f"  Type: {item['algorithm']}")
                print(f"  Plain: {item['plaintext']}")
                print()
        else:
            print("## No hashes cracked")


def main():
    if len(sys.argv) < 2:
        print("BLACK-HATER Hash Cracker")
        print("\nUsage:")
        print("  python3 hash_cracker.py <hash>")
        print("  python3 hash_cracker.py --file <hashes.txt>")
        print("  python3 hash_cracker.py --type md5 --wordlist rockyou.txt <hash>")
        print("\nExamples:")
        print("  python3 hash_cracker.py 5f4dcc3b5aa765d61d8327deb882cf99")
        print("  python3 hash_cracker.py --type sha256 --wordlist rockyou.txt <hash>")
        sys.exit(1)
    
    cracker = HashCracker()
    
    # Parse arguments
    hash_str = None
    algorithm = None
    wordlist = None
    
    args = sys.argv[1:]
    i = 0
    while i < len(args):
        if args[i] == '--type' and i + 1 < len(args):
            algorithm = args[i + 1]
            i += 2
        elif args[i] == '--wordlist' and i + 1 < len(args):
            wordlist = args[i + 1]
            i += 2
        elif args[i] == '--file' and i + 1 < len(args):
            # TODO: Handle file input
            print("[ERROR] File input not yet implemented")
            sys.exit(1)
        else:
            hash_str = args[i]
            i += 1
    
    if not hash_str:
        print("[ERROR] No hash provided")
        sys.exit(1)
    
    # Identify hash type
    if not algorithm:
        possible = cracker.identify_hash(hash_str)
        print(f"[*] Possible hash types: {', '.join(possible)}")
        algorithm = possible[0]
    
    print(f"[*] Algorithm: {algorithm}")
    print(f"[*] Hash: {hash_str}\n")
    
    # Try to crack
    result = None
    
    if wordlist:
        result = cracker.crack_wordlist(hash_str, algorithm, wordlist)
    
    if not result:
        # Try common passwords first
        common_passwords = [
            'password', '123456', 'admin', 'root', 'test',
            'letmein', 'welcome', 'monkey', 'dragon', 'master',
            'qwerty', 'login', 'abc123', 'password1', '1234567890',
        ]
        
        print("[*] Trying common passwords...")
        for word in common_passwords:
            hash_func = cracker.ALGORITHMS.get(algorithm)
            if hash_func:
                computed = hash_func(word.encode()).hexdigest()
                if computed == hash_str.lower():
                    result = word
                    print(f"[+] FOUND: {word}")
                    break
    
    if not result and not wordlist:
        # Brute force with lowercase + digits
        result = cracker.crack_bruteforce(hash_str, algorithm, max_length=4)
    
    if result:
        cracker.cracked.append({
            'hash': hash_str,
            'algorithm': algorithm,
            'plaintext': result
        })
    
    cracker.generate_report()


if __name__ == '__main__':
    main()
