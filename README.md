# Synopsis

Run the ping thing.

```
$ docker build --rm -t up-or-not upornot
$ docker run -it --rm -p 4444:4444 upornot 1.1.1.1
```

Update the URL in `main.py` and run the checker on a raspberry pi:

```
$ python3 main.py
```
