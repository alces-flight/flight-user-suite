#!/usr/bin/env python3

import pam
import sys

if __name__ == "__main__":
    if len(sys.argv) != 3:
        print("Usage: {} <username> </path/to/fifo>".format(sys.argv[0]))
        sys.exit(1)
    username = sys.argv[1]
    fifo_path = sys.argv[2]
    with open(fifo_path) as fifo:
        password = fifo.read()
        success = pam.authenticate(username, password, service="login")
        if success:
            sys.exit(0)
        else:
            sys.exit(1)
