#!/bin/bash
# @description Install pingmonke as a launchd service

PLIST="/Library/LaunchDaemons/com.pingmonke.service.plist"
cat <<EOF > $PLIST
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key><string>com.pingmonke.service</string>
    <key>ProgramArguments</key><array><string>/usr/local/bin/pingmonke</string></array>
    <key>RunAtLoad</key><true/>
</dict>
</plist>
EOF

launchctl load $PLIST