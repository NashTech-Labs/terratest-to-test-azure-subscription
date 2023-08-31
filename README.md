# Terratest to test the complete Azure subscription 
 

## This repo contains the terratest  test cases for the Azure subscription.

 


-------------
### To run this terratest, You must have the terraform subscription module and should be in the same root and you need to export the given variale in your environment and then run go test.

 


1. You have to export the credential of azure.

 

        export CLIENT_ID=""

 

        export CLIENT_SECRET=""

 

        export TENANT_ID=""

 

        export SUBSCRIPTION_ID=""%           


2. Add the following values in the terratest code:


        	subscription_name = "< >"

            billing_account_name = "< >"

            workload = "< >"

            enrollment_account_name = "< >"

            managementgroupassociation = "< >"

3. In the last, you need to run the below command to run the test case:-

             go mod init <>

             go mod tidy

             go test -v