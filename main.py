from gpiozero import PWMLED
from time import sleep

# +------------------| |--| |------+
# | ooooooooooooo P1 |C|  |A|      |
# | 1oooooooooooo    +-+  +-+      |
# |    1ooo                        |
# | P5 oooo        +---+          +====
# |                |SoC|          | USB
# |   |D| Pi Model +---+          +====
# |   |S| B  V2.0                  |
# |   |I|                  |C|+======
# |                        |S||   Net
# |                        |I|+======
# =pwr             |HDMI|          |
# +----------------|    |----------+
# 
# 
# P1:
#    3V3  (1) (2)  5V    
#  GPIO2  (3) (4)  5V    
#  GPIO3  (5) (6)  GND   
#  GPIO4  (7) (8)  GPIO14
#    GND  (9) (10) GPIO15
# GPIO17 (11) (12) GPIO18
# GPIO27 (13) (14) GND   
# GPIO22 (15) (16) GPIO23
#    3V3 (17) (18) GPIO24
# GPIO10 (19) (20) GND   
#  GPIO9 (21) (22) GPIO25
# GPIO11 (23) (24) GPIO8 
#    GND (25) (26) GPIO7 

leds = {
  # Pin 11 / GPIO17
  "green": PWMLED(17),
  # Pin 12 / GPIO18
  "red": PWMLED(18)
}

def on(name):
  print("%s: turn on" % name)
  led = leds[name]
  val = led.value
  while led.value < 0.9:
    val = val + (1.0 - val) / 2.0
    led.value = val
    print("... %0.2f" % val)
    sleep(1)
  led.on()

def off(name):
  print("%s: turn off" % name)
  led = leds[name]
  val = led.value
  while led.value > 0.1:
    val = val / 2.0
    led.value = val
    print("... %0.2f" % val)
    sleep(1)
  led.off()

# curl http://up-or-not.pickardayune.com:8080/api/status
on("green")
on("red")
off("green")
off("red")
