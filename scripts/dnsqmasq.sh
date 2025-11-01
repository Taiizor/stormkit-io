#!/usr/bin/env bash

# In order to run this script use:
# $ bash scripts/dnsmasq.sh

# Install dnsmasq
brew install dnsmasq

# Create config folder if it doesn’t already exist
mkdir -pv $(brew --prefix)/etc/

# Configure dnsmasq for *.stormkit
echo 'address=/.stormkit/127.0.0.1' >> $(brew --prefix)/etc/dnsmasq.conf

# Configure the port for macOS High Sierra
echo 'port=53' >> $(brew --prefix)/etc/dnsmasq.conf

# Start dnsmasq as a service so it automatically starts at login
sudo brew services start dnsmasq

# Create a dns resolver
sudo mkdir -pv /etc/resolver
sudo bash -c 'echo "nameserver 127.0.0.1" > /etc/resolver/stormkit'

# Verify that all .sk.local requests are using 127.0.0.1
scutil --dns
