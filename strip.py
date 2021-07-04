# first:
# sudo pip3 install rpi_ws281x adafruit-circuitpython-neopixel
# sudo python3 -m pip install --force-reinstall adafruit-blinka

# this script must be run as root.

# (edge of board)
#    3V3  (1) (2)  5V    
#  GPIO2  (3) (4)  5V    
#  GPIO3  (5) (6)  GND     <----- ground
#  GPIO4  (7) (8)  GPIO14
#    GND  (9) (10) GPIO15
# GPIO17 (11) (12) GPIO18  <----- data
# GPIO27 (13) (14) GND   
# GPIO22 (15) (16) GPIO23
#    3V3 (17) (18) GPIO24
# GPIO10 (19) (20) GND   
#  GPIO9 (21) (22) GPIO25
# GPIO11 (23) (24) GPIO8 
#    GND (25) (26) GPIO7 
# (more space, then usb plugs)

import board
import neopixel
from time import sleep
import json
from urllib.request import urlopen

#p = {}
p = neopixel.NeoPixel(board.D18, 95, brightness = 1)

MILLI = 1000000
BEST_NANOS  =  20*MILLI
WORST_NANOS = 100*MILLI

def run(url):
  max = 0
  while True:
    try:
      response = urlopen(url)
      data = json.loads(response.read())
      recent = data["recent"]
      recent.reverse()
      for i, r in enumerate(recent):
        if i > max:
          max = i
        if r["status"] != "OK":
          p[i] = (255,0,0)
        else:
          nanos = r["nanos"]
          if nanos < BEST_NANOS:
            p[i] = (0,0,255) # RBG
          elif nanos > WORST_NANOS:
            p[i] = (255,0,0)
          else:
            pos = 255 * (nanos - BEST_NANOS) / (WORST_NANOS - BEST_NANOS)
            p[i] = (pos, 0, 255-pos)
    except (IOError, ValueError) as e:
      p[0] = (0,255,0)
      for i in range(max):
        p[i+1] = (0,0,0)
      # If tiny isn't accessible, assume the internet is DOWN
      print(e)
    #print(p)
    sleep(1)

try:
  run("http://up-or-not.pickardayune.com:8080/api/target/8.8.8.8/recent")
except KeyboardInterrupt:
  p.deinit()
  print("exit.")
