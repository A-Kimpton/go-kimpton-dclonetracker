echo "Building discord bot binary server.exe..."
go build .\cmd\discordbot\server.go
echo "complete server.exe"

echo "" 
echo ""

echo "Building d2io client binary client.exe..."
go build .\cmd\cli\client.go
echo "complete client.exe"