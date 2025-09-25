## Plane
TCP forward.

## Configs
将`127.0.0.1:8080`和`127.0.0.1:8081`反向代理到`7000`端口。

将`127.0.0.1:3306``反向代理到`9306`端口。
```json
[
  {
    "listener_port": 7000,
    "forward": [
      "127.0.0.1:8080",
      "127.0.0.1:8081"
    ]
  },
  {
    "listener_port": 9306,
    "forward": [
      "127.0.0.1:3306"
    ]
  }
]
```

## Use
```shell
go build
```
```shell
./plane 
```
or
```shell
./plane -f config.json
```