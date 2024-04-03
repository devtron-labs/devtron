echo "creating binary file..."
chmod 777 WiringNilCheck.go
go build -o wirenilcheckbinary .
chmod 557 WiringNilCheck.go
echo "binary file created"
ls | grep wirenilcheckbinary #  checking whether binary file is created or not
./wirenilcheckbinary
