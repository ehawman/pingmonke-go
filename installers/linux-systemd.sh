#!/bin/bash
# @description Install pingmonke as a systemd service

SERVICE_NAME="pingmonke"
BIN_PATH="/usr/local/bin/pingmonke"

echo "[Unit]
Description=Pingmonke Service
After=network.target

[Service]
ExecStart=$BIN_PATH
Restart=always

[Install]
WantedBy=multi-user.target" > /etc/systemd/system/$SERVICE_NAME.service

systemctl daemon-reload
systemctl enable $SERVICE_NAME
systemctl start $SERVICE_NAME