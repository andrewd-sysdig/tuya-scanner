# tuya-scanner — Scanner with MQTT and Prometheus exporter for the TUYA Devices

# buid
```
GOOS=linux GOARCH=arm GOARM=7 go build -o tuya-scanner ./cmd/worker/
```

# create
```
sudo mkdir /opt/tuya-scanner
sudo chmod +x /opt/tuya-scanner
```

- create scanner.env 
```
LISTENING_ADDRESS=0.0.0.0:9265
BROKER_ADDRESS=mqtt.server.local
BROKER_PORT=1883
BROKER_USER=user
BROKER_PASSWORD=password
BROKER_TOPIC=tuya-scanner
CLIENT_ID=tuya-scanner
```

- create devices.json
```json
[
  {
    "gwId": "xxxxxxxxxxxxxxxxxxxxxx",
    "key": "xxxxxxxxxxxxxxxx",
    "type": "Switch",
    "name": "test_device_1"
  },
  {
    "gwId": "xxxxxxxxxxxxxxxxxxxxxx",
    "key": "xxxxxxxxxxxxxxxx",
    "type": "Switch",
    "name": "test_device_2"
  }
]
```

- create service
sudo nano /etc/systemd/system/tuya-scanner.service

```
[Unit]
Description=tuya-scanner
Documentation=https://github.com/sysdigdan/tuya-scanner
After=network-online.target

[Service]
User=pi
Restart=on-failure
WorkingDirectory=/opt/tuya-scanner
ExecStartPre=/bin/sleep 15
ExecStart=/opt/tuya-scanner/tuya-scanner

[Install]
WantedBy=multi-user.target
```

# run
---------------------------
sudo systemctl enable tuya-scanner
sudo systemctl start tuya-scanner
sudo systemctl status tuya-scanner