import os
import subprocess as sp

def printHeading(heading):
	print("-"*25)
	print(heading)
	print("-"*25)

# printHeading("Printing the Current Environment Variables")
# os.system("printenv")

print("")
printHeading("Extracting Passwords...")

Pg_User="postgres"
Pg_Password=sp.getoutput("kubectl -n devtroncd get secret devtron-secret -o jsonpath='{.data.PG_PASSWORD}' | base64 -D")
Acd_Username="admin"
Acd_Password=sp.getoutput("kubectl -n devtroncd get secret devtron-secret -o jsonpath='{.data.ACD_PASSWORD}' | base64 -D")
Cloud_Provider="MINIO"
EndpointStore=sp.getoutput("kubectl get ep -ndevtroncd -l app=minio -o jsonpath='{.items[0].subsets[0].addresses[0].ip}'")
Minio_Endpoint="http://"+EndpointStore+":9000/minio"
Minio_Access_Key=sp.getoutput("kubectl -n devtroncd get secret devtron-secret -o jsonpath='{.data.MINIO_ACCESS_KEY}' | base64 -D")
Minio_Secret_Key=sp.getoutput("kubectl -n devtroncd get secret devtron-secret -o jsonpath='{.data.MINIO_SECRET_KEY}' | base64 -D")

print("")
print("Database Username :",Pg_User)
print("Database Password :",Pg_Password)
print("ArgoCD Username :",Acd_Username)
print("ArgoCD Password :",Acd_Password)
print("Cloud Provider :",Cloud_Provider)
print("Minio Endpoint :",Minio_Endpoint)
print("Minio Access Key :",Minio_Access_Key)
print("Minio Secret Key :",Minio_Secret_Key)

printHeading("Setting up the Environment Varibales")
os.system("sleep 3")
os.environ['PG_USER']=Pg_User
os.environ['PG_PASSWORD']=Pg_Password
os.environ['ACD_USERNAME']=Acd_Username
os.environ['ACD_PASSWORD']=Acd_Password
os.environ['CLOUD_PROVIDER']=Cloud_Provider
os.environ['MINIO_ENDPOINT']=Minio_Endpoint
os.environ['MINIO_ACCESS_KEY']=Minio_Access_Key
os.environ['MINIO_SECRET_KEY']=Minio_Secret_Key

print("")
print("Environment Variables has been Exported")
print("")

print("Executing Devtron Binary..")
os.system("sleep 3")

# printHeading("After Exporting Values")
# os.system("printenv")

os.system("./devtron")
