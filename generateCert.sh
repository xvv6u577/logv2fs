#!/bin/sh

# 1. Generate CA private key and self-signed certificate
openssl req -x509 -newkey rsa:4096 -nodes -days 365 -keyout ./CA/ca-key.pem -out ./CA/ca-cert.pem -subj "/C=HK/ST=ASIA/L=HONGKONG/O=DEV/OU=TUTORIAL/CN=*.undervineyard.com"

# 2. Generate Web Server’s Private Key and CSR (Certificate Signing Request)
openssl req -newkey rsa:4096 -nodes -keyout ./CA/server-key.pem -out ./CA/server-req.pem -subj "/C=HK/ST=ASIA/L=HONGKONG/O=DEV/OU=TUTORIAL/CN=*.undervineyard.com"

# 3. Sign the Web Server Certificate Request (CSRE
openssl x509 -req -in ./CA/server-req.pem -CA ./CA/ca-cert.pem -CAkey ./CA/ca-key.pem -CAcreateserial -out ./CA/server-cert.pem -extfile ./CA/server-extra.conf

# 4. Generate client’s private key and certificate signing request (CSR)
openssl req -newkey rsa:4096 -nodes -keyout ./CA/client-key.pem -out ./CA/client-req.pem -subj "/C=HK/ST=ASIA/L=HONGKONG/O=DEV/OU=TUTORIAL/CN=*.undervineyard.com"

# 5. Sign the Client Certificate Request (CSR)
openssl x509 -req -in ./CA/client-req.pem -days 60 -CA ./CA/ca-cert.pem -CAkey ./CA/ca-key.pem -CAcreateserial -out ./CA/client-cert.pem -extfile ./CA/client-extra.conf