#!/bin/bash

# Forward TCP traffic on port 8000 to port 8080 on the eth0 interface.
iptables -t nat -A PREROUTING -p tcp -i eth0 --dport 8000 -j REDIRECT --to-port 8080

# List all iptables rules.
iptables -t nat --list
