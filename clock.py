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
import time

#p = {}
p = neopixel.NeoPixel(board.D18, 95, brightness = 1)

def run():
  while True:
    t = int(time.time())
    i = 0
    while t > 0:
      if t & 0x01 == 0:
        p[i] = (255,0,0)
      else:
        p[i] = (0,255,0)
      t = int(t / 2)
      i = i + 1
    time.sleep(0.3)

try:
  run()
except KeyboardInterrupt:
  p.deinit()
  print("exit.")
