#!/bin/bash

# TODO: Log user activity to stdout, date + time when a user is active on specific pages
# Pages:
#       - /
#       - /feed
#       - /profile
#       - /editor
#       - /help
#       - /trouble-reports
#       - /notes
#       - /tools
#
# `✅ 2025/10/20 23:52:32 [Server] 200 GET     /pg-press/ (92.206.35.156) 3.10982ms User{ID: 1268219808, Name: knackwurstking (admin) [has API key]}
`
# `✅ 2025/10/20 23:53:03 [Server] 200 GET     /pg-press/trouble-reports (92.206.35.156) 1.518576ms User{ID: 1268219808, Name: knackwurstking (admin) [has API key]}
`
#
# Log: date + time string, user ID + name, server path
#
# NOTE: Try to read the logs from the `journalctl --user -u pg-press --output cat` command somehow
