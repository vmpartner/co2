# Watcher based on MH-Z19 carbon dioxide gas sensor
Watcher written on golang, its watch PPM from sensor and send alert to telegram if value increasing.
Users who receiving alerts maybe more than one.  

I use arduino mega 2560 + MH-Z19 sensor. You can get my example in folder "sketch". 

## Start
1. Prepare your sensor, see sketch folder
2. Rename app.example.conf > app.conf and set your [telegram token](https://core.telegram.org/bots), user, com/port of sensor
3. go run main.go

## Command telegram bot
/start - start watching  
/stop - stop watching  
/sleep 120 - sleep 120 minutes (120 is example)
