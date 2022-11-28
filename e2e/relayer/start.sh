#!/bin/bash

set -e

# restore keys
rly keys restore cosmos gaia-key "reopen throw concert garment wash slab jump company habit father below stage float attitude achieve net charge bulb mouse mind fat net hello vague"
rly keys restore juno juno-key "predict lonely oxygen category agent wait legend quarter often six liquid search panic panther chuckle glow alone detail bike below dust marriage throw pause"

# create client, connection and transfer channels
rly paths new gaia juno test_path
rly transact link test_path

# start relayer
rly start test_path -p events
