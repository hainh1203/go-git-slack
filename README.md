# Setup
Config
```shell
cp config.json.example config.json
```

Instructions for creating GoogleSheetApi (config.json)
```
https://github.com/hainh1203/go-git-slack/blob/main/setup-google-sheet-api/README.md
```

# Dev
Run with port default (8080)
```shell
go run main.go
```
Custom port (ex: 9999)
```shell
go run main.go 9999
```

# Prod
Compile the code into an executable
```shell
go build main.go
```
Run with port default (8080)
```shell
./main
```
Custom port (ex: 9999)
```shell
./main 9999
```

# Prod with docker

Build image
```
docker build -t go-git-slack .
```
Run with port 9999
```
docker run -dp 9999:8080 -v $PWD/config.json:/app/config.json --name go-git-slack go-git-slack
```
Access container
```
docker exec -it go-git-slack bash
```

# Endpoint for gitlab
```
/gitlab
```