#!/usr/bin/env python3
"""
TAMK Asset Decryptor - Decrypt T.A.M.K encrypted assets.
Part of T.A.M.K Android Web Pentester skill.

Usage:
    python3 asset_decryptor.py <encrypted-file> <password>
    python3 asset_decryptor.py --batch <directory> <password>
"""

import os
import sys
import hashlib
import struct
from pathlib import Path
from typing import Optional, Tuple

# Try to import cryptography
try:
    from cryptography.hazmat.primitives.ciphers import Cipher, algorithms, modes
    from cryptography.hazmat.primitives import padding
    from cryptography.hazmat.backends import default_backend
    HAS_CRYPTOGRAPHY = True
except ImportError:
    HAS_CRYPTOGRAPHY = False


# TAMK Encryption constants
MAGIC_HEADER = b'TAMK_ENC_1'
SALT_SIZE = 16
IV_SIZE = 12  # AES-GCM IV


class TAMKDecryptor:
    """Decrypt T.A.M.K encrypted assets."""
    
    def __init__(self, password: str):
        self.password = password.encode('utf-8')
    
    def derive_key(self, salt: bytes) -> bytes:
        """Derive AES key from password and salt using SHA-256."""
        hasher = hashlib.sha256()
        hasher.update(self.password)
        hasher.update(salt)
        return hasher.digest()[:32]  # AES-256
    
    def decrypt_file(self, input_path: str, output_path: Optional[str] = None) -> bool:
        """Decrypt a TAMK encrypted file."""
        input_path = Path(input_path)
        
        if not input_path.exists():
            print(f"[ERROR] File not found: {input_path}")
            return False
        
        # Read encrypted data
        with open(input_path, 'rb') as f:
            data = f.read()
        
        # Verify magic header
        if data[:len(MAGIC_HEADER)] != MAGIC_HEADER:
            print(f"[ERROR] Not a TAMK encrypted file (invalid magic header)")
            return False
        
        if len(data) < len(MAGIC_HEADER) + SALT_SIZE + IV_SIZE:
            print(f"[ERROR] File too small to be encrypted")
            return False
        
        # Extract components
        offset = len(MAGIC_HEADER)
        salt = data[offset:offset + SALT_SIZE]
        offset += SALT_SIZE
        
        iv = data[offset:offset + IV_SIZE]
        offset += IV_SIZE
        
        ciphertext = data[offset:]
        
        # Derive key
        key = self.derive_key(salt)
        
        # Decrypt
        if HAS_CRYPTOGRAPHY:
            return self._decrypt_with_cryptography(key, iv, ciphertext, input_path, output_path)
        else:
            return self._decrypt_without_cryptography(key, iv, ciphertext, input_path, output_path)
    
    def _decrypt_with_cryptography(self, key: bytes, iv: bytes, ciphertext: bytes,
                                    input_path: Path, output_path: Optional[str]) -> bool:
        """Decrypt using cryptography library (AES-256-GCM)."""
        try:
            cipher = Cipher(
                algorithms.AES(key),
                modes.GCM(iv),
                backend=default_backend()
            )
            decryptor = cipher.decryptor()
            plaintext = decryptor.update(ciphertext) + decryptor.finalize()
            
            # Determine output path
            if output_path is None:
                output_path = input_path.with_suffix('')
            
            # Write decrypted file
            with open(output_path, 'wb') as f:
                f.write(plaintext)
            
            print(f"[OK] Decrypted: {input_path.name} -> {output_path}")
            return True
            
        except Exception as e:
            print(f"[ERROR] Decryption failed: {e}")
            return False
    
    def _decrypt_without_cryptography(self, key: bytes, iv: bytes, ciphertext: bytes,
                                       input_path: Path, output_path: Optional[str]) -> bool:
        """Decrypt without cryptography library (simple XOR for demonstration)."""
        print("[WARN] cryptography library not available, using XOR fallback")
        print("[WARN] Install with: pip install cryptography")
        
        # XOR decryption (NOT secure, just for demonstration)
        key_extended = (key * (len(ciphertext) // len(key) + 1))[:len(ciphertext)]
        plaintext = bytes(a ^ b for a, b in zip(ciphertext, key_extended))
        
        # Determine output path
        if output_path is None:
            output_path = input_path.with_suffix('')
        
        # Write decrypted file
        with open(output_path, 'wb') as f:
            f.write(plaintext)
        
        print(f"[OK] Decrypted (XOR): {input_path.name} -> {output_path}")
        return True
    
    def batch_decrypt(self, directory: str) -> Tuple[int, int]:
        """Decrypt all .enc files in a directory."""
        directory = Path(directory)
        
        if not directory.exists():
            print(f"[ERROR] Directory not found: {directory}")
            return 0, 0
        
        success_count = 0
        fail_count = 0
        
        for enc_file in directory.rglob("*.enc"):
            if self.decrypt_file(str(enc_file)):
                success_count += 1
            else:
                fail_count += 1
        
        return success_count, fail_count
    
    def try_passwords(self, input_path: str, passwords: list) -> Optional[str]:
        """Try multiple passwords to decrypt a file."""
        for password in passwords:
            self.password = password.encode('utf-8')
            
            # Read encrypted data
            with open(input_path, 'rb') as f:
                data = f.read()
            
            # Extract components
            offset = len(MAGIC_HEADER)
            salt = data[offset:offset + SALT_SIZE]
            offset += SALT_SIZE
            iv = data[offset:offset + IV_SIZE]
            offset += IV_SIZE
            ciphertext = data[offset:]
            
            # Derive key
            key = self.derive_key(salt)
            
            # Try decryption
            if HAS_CRYPTOGRAPHY:
                try:
                    cipher = Cipher(
                        algorithms.AES(key),
                        modes.GCM(iv),
                        backend=default_backend()
                    )
                    decryptor = cipher.decryptor()
                    plaintext = decryptor.update(ciphertext) + decryptor.finalize()
                    
                    # Check if decryption seems successful
                    if len(plaintext) > 0:
                        print(f"[OK] Password found: {password}")
                        return password
                except Exception:
                    continue
        
        return None


def main():
    if len(sys.argv) < 2:
        print("TAMK Asset Decryptor")
        print("\nUsage:")
        print("  python3 asset_decryptor.py <encrypted-file> <password>")
        print("  python3 asset_decryptor.py --batch <directory> <password>")
        print("  python3 asset_decryptor.py --try <encrypted-file> <password1,password2,...>")
        print("\nExamples:")
        print("  python3 asset_decryptor.py index.html.enc mypassword")
        print("  python3 asset_decryptor.py --batch ./assets/ mypassword")
        print("  python3 asset_decryptor.py --try index.html.enc pass1,pass2,pass3")
        sys.exit(1)
    
    if sys.argv[1] == '--batch':
        # Batch decrypt mode
        if len(sys.argv) < 4:
            print("[ERROR] Usage: --batch <directory> <password>")
            sys.exit(1)
        
        directory = sys.argv[2]
        password = sys.argv[3]
        
        decryptor = TAMKDecryptor(password)
        success, fail = decryptor.batch_decrypt(directory)
        
        print(f"\n[SUMMARY] Decrypted: {success}, Failed: {fail}")
        
    elif sys.argv[1] == '--try':
        # Try passwords mode
        if len(sys.argv) < 4:
            print("[ERROR] Usage: --try <encrypted-file> <password1,password2,...>")
            sys.exit(1)
        
        input_path = sys.argv[2]
        passwords = sys.argv[3].split(',')
        
        decryptor = TAMKDecryptor('')
        found = decryptor.try_passwords(input_path, passwords)
        
        if found:
            # Now decrypt with the found password
            decryptor.password = found.encode('utf-8')
            decryptor.decrypt_file(input_path)
        else:
            print("[FAIL] No working password found")
            
    else:
        # Single file decrypt
        if len(sys.argv) < 3:
            print("[ERROR] Usage: asset_decryptor.py <encrypted-file> <password>")
            sys.exit(1)
        
        input_path = sys.argv[1]
        password = sys.argv[2]
        
        decryptor = TAMKDecryptor(password)
        decryptor.decrypt_file(input_path)


if __name__ == '__main__':
    main()
