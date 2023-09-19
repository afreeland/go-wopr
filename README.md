# go-wopr

War Operation Plan Response (WOPR)  
🔊 pronounced "whopper"

If you live under a 🪨 you can learn more about [WarGames here](https://en.wikipedia.org/wiki/WarGames)

Introduced this movie to my 13 year old son friday night and thought that it would make a good saturday evening project =)

Releases
---
You can find downloads here  
https://github.com/afreeland/go-wopr/releases/tag/wopr  

Flavors available:
- darwin-amd64  
```
curl -L https://github.com/afreeland/go-wopr/releases/download/wopr/wopr-wopr-darwin-amd64.tar.gz | tar -xz
```
- darwin-arm64
```
curl -L https://github.com/afreeland/go-wopr/releases/download/wopr/wopr-wopr-darwin-arm64.tar.gz | tar -xz
```
- linux-386
```
curl -L https://github.com/afreeland/go-wopr/releases/download/wopr/wopr-wopr-linux-386.tar.gz | tar -xz
```
- linux-amd64
```
curl -L https://github.com/afreeland/go-wopr/releases/download/wopr/wopr-wopr-linux-amd64.tar.gz | tar -xz
```
- linux-arm64
```
curl -L https://github.com/afreeland/go-wopr/releases/download/wopr/wopr-wopr-linux-arm64.tar.gz | tar -xz
```
- windows-386
```
curl -LJO https://github.com/afreeland/go-wopr/releases/download/wopr/wopr-wopr-windows-386.zip
```
- windows-amd64
```
curl -LJO https://github.com/afreeland/go-wopr/releases/download/wopr/wopr-wopr-windows-amd64.zip
```


# Running 

There are currently two ways in which you can use this

### Console Mode
This allows you to interact with it from command line by simply running the binary  
```
./wopr
```  

### Networking Mode
This allows you to interact with it by actually connecting to it over TCP port 2000 by default.

This can be done by running in `server` mode like so
```
./wopr --server
```

Then open another terminal and utilize something like netcat or telnet.

```
nc localhost 2000
```

## How To
The WOPR will go through its self identification phase and then prompt user with a logon

- You may run things like `help` `help logon`
- You may get _disconnected_ if auth is not correct
- You may `help games` followed by `list games`
- Ultimately you will want to logon, `joshua` may be able to help


Since the movie is scripted there are some parts where user input is not necessarily relevant.  I may later intergrate with a guided ChatGPT setup...for more relevant up to date kind of interactions.


Inspiration
----

[natebeck/go-wopr](https://github.com/natebeck/go-wopr/blob/master/wopr.go) - A nice repo on that utilized TCP connections to simulate how the interactions worked in the movie.

[linuxlawson/wargames](https://github.com/linuxlawson/wargames) - A python implementation which is easily ran and fun to play

[ASCII Art](http://www.48k.ca/wgascii.html) - Has a really good breakdown and sample of the ASCII art for creating the US and Soviet Union.

[Font](https://www.fontstruct.com/fontstructions/show/1616604/wopr-terminal) - Font to match the movie done by [Lord Nightmare](https://www.fontstruct.com/fontstructors/59995/lord_nightmare)


TODO
-----

- Websocket/html implementation for anyone wanting to host
- Sound effects to match movie "Shall we play a game?"
- ChatGPT variant for chatting with.  Could be semi-scripted or prompt that is "aware of game"


Buid Dockerfile
-----

### From Docker Hub
#### pull from docker hub
```
docker pull afreeland/go-wopr@latest
```
#### run in console mode
```
docker run -it afreeland/go-wopr@latest
```

#### run in network mode
```
docker run afreeland/go-wopr@latest --server
```

### From Code

#### build from code
```
# Build the Docker image with the default port (2000)
docker build -t wopr .
```

#### run as console
```
# Run the Docker container with the default port (2000) and without the --server flag
docker run -it wopr
```

### run as server
```
docker run -p 2000:2000 wopr --server
```
