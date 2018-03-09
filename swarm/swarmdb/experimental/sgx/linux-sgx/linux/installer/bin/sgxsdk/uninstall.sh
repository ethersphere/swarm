#!/usr/bin/env bash

# Removing the SDK folder
rm -fr /var/www/vhost/sgx/linux-sgx/linux/installer/bin/sgxsdk 2> /dev/null

if [ $? -ne 0 ]; then
    echo "Superuser privilege is required."
    exit 1
fi

echo "Intel(R) SGX SDK uninstalled."

