chmod 777 WiringNilCheck.go
sed -i 's/func CheckIfNilInWire()/func main()/1' WiringNilCheck.go
sed -i 's/func main()/func tempMain()/1' main.go
sed '1i //go:build ignore' Wire.go
sed '1i //go:build ignore' main.go
sed -i 's/func InitializeApp()/func InitializeApp1()/1' Wire.go
go build -o wirenilcheckbinary *.go
sed -i 's/func InitializeApp1()/func InitializeApp()/1' Wire.go
sed '1,2d' main.go
sed '1,2d' Wire.go
sed -i 's/func tempMain()/func main()/1' main.go
sed -i 's/func main()/func CheckIfNilInWire()/1' WiringNilCheck.go
chmod 557 WiringNilCheck.go
mv wirenilcheckbinary ./tests/integrationTesting
cd ./tests/integrationTesting
ls | grep wirenilcheckbinary #  checking whether binary file is created or not