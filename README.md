# Watcher based on MH-Z19 carbon dioxide gas sensor
Watcher written on golang, its watch PPM from sensor and send alert to telegram if value increasing.
Users who receiving alerts maybe more than one.

## Start
1. Prepare your sensor, see sketch folder
2. Rename app.example.conf > app.conf and set your [telegram token](https://core.telegram.org/bots) and com/port of sensor
3. go run main.go

## Command telegram bot
/start - start watching  
/stop - stop watching  
/sleep 120 - sleep 120 minutes  
