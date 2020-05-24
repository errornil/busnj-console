# busnj-console

Server for [BusNJ Console](https://console.busnj.chuhlomin.com).

# Quickstart

```
$ make build-static build-docker run-docker
Sending build context to Docker daemon  7.504MB
Step 1/3 : FROM scratch
 ---> 
Step 2/3 : ADD server /server
 ---> 69655bab0f76
Step 3/3 : CMD ["/server"]
 ---> Running in 27215179d230
Removing intermediate container 27215179d230
 ---> f1723ac81a29
Successfully built f1723ac81a29
Successfully tagged busnj-console:latest
2019/05/14 01:24:41 Starting...
2019/05/14 01:24:41 Listening on port 6001...
```
