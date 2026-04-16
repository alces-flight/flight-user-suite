#!/usr/bin/env python3

import pam
import sys

if __name__ == "__main__":
    if len(sys.argv) != 2:
        print("Usage: {} <username>".format(sys.argv[0]))
        print("  Password is read from standard input")
        sys.exit(1)
    username = sys.argv[1]
    password = sys.stdin.read().strip()
    success = pam.authenticate(username, password, service="login")
    if success:
        sys.exit(0)
    else:
        sys.exit(1)
