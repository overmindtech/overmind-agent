#!/bin/bash

# Simulate logging in to a very slow machine...
echo "Hi!"
sleep 0.2
echo "Welcome to ABC Corp..."
sleep 0.2
echo "Users must be authenticated!"
sleep 0.2

# Read Password
echo -n "Password: "
read -s password
echo

# Write password
echo $password
