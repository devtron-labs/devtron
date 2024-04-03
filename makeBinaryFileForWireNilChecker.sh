echo "creating binary file..."
cp auth_model.conf ./tests/integrationTesting
chmod 777 WiringNilCheck.go
go build -o wirenilcheckbinary .
chmod 557 WiringNilCheck.go
mv wirenilcheckbinary ./tests/integrationTesting
cd tests/integrationTesting
echo "binary file $(ls | grep wirenilcheckbinary) is created"
