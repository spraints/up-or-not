from gpiozero import PWMLED
from time import sleep
import urllib.request, json

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

def run(url):
  while True:
    try:
      response = urllib.request.urlopen(url)
      data = json.loads(response.read())
      green_score = score_green(data)
      red_score = score_red(data)
      print("ok %0.2f / bad %0.2f" % (green_score, red_score))
      leds["green"].value = green_score
      leds["red"].value = red_score
    except (IOError, ValueError) as e:
      # If tiny isn't accessible, assume the internet is DOWN
      print(e)
      leds["green"].value = 0.0
      leds["red"].value = 1.0
    sleep(10)

def score_green(data):
  score = 0.0
  factor = 1.0
  for bucket in data["buckets"]:
    score += factor * bucket["count"]
    factor /= 2
  return score / data["count"]

def score_red(data):
  score = 0.0
  score += (data["count"] - data["ok"])
  factor = 1.0
  # start at the end, but only look at the worst two buckets.
  for bucket in data["buckets"][::-1][0:2]:
    score += factor * bucket["count"]
    factor /= 2
  return score / data["count"]

try:
  run("http://up-or-not.pickardayune.com:8080/api/target/8.8.8.8")
except KeyboardInterrupt:
  for name in leds:
    leds[name].off()
  print("exit.")
