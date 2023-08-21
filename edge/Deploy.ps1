$env:GOOS = 'linux'
$env:GOARCH = 'arm'
New-Item -ItemType Directory -Path build
go build -o build\core-gateway core-gateway\main.go
go build -o build\core-bus core-bus\main.go
#go build -o build\proto-bee proto-bee\main.go
pscp -pw fa .\build\* root@192.168.1.5:/opt/cobra
plink.exe -ssh -t -pw fa root@192.168.1.5  -m plink-script.txt
