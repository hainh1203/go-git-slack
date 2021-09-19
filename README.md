Config
```shell
cp config.json.example config.json
```

Sau khi đã tạo file config.json, bước tiếp theo cần thiết lập GoogleSheetApi
```
https://github.com/hainh1203/go-git-slack/blob/main/setup-google-sheet-api/README.md
```

Dev
```shell
go run main.go
```

Build
```shell
go build main.go
./main
```

Prod
```
docker build -t go-git-slack .
docker run -dp 9999:9999 go-git-slack
```

Gitlab
```
http://localhost:9999/gitlab
```