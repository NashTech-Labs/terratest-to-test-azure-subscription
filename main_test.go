package main

import (
	// "context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/stretchr/testify/assert"
)

var (
	globalBackendConf          = make(map[string]interface{})
	globalEnvVars              = make(map[string]string)
	subscription_name          = "< >"
	billing_account_name       = "< >"
	workload                   = "< >"
	enrollment_account_name    = "< >"
	managementgroupassociation = "< >"
)

type Subscription struct {
	ID          string `json:"id"`
	DisplayName string `json:"displayName"`
	State       string `json:"state"`
	TenantID    string `json:"tenantId"`
}

const (
	apiVersion = "2020-09-01"
)

func setTerraformVariables() (map[string]string, error) {

	CLIENT_ID := os.Getenv("CLIENT")
	CLIENT_SECRET := os.Getenv("CLIENT_SECRET_ID")
	TENANT_ID := os.Getenv("TENANT")
	SUBSCRIPTION_ID := os.Getenv("SUBSCRIPTION")

	fmt.Println("##################################Environment Variables:#######################################")
	fmt.Println("AZURE_CLIENT_ID:", CLIENT_ID)
	fmt.Println("AZURE_CLIENT_SECRET:", CLIENT_SECRET)
	fmt.Println("AZURE_TENANT_ID:", TENANT_ID)
	fmt.Println("AZURE_SUBSCRIPTION_ID:", SUBSCRIPTION_ID)
	fmt.Println("****************************************Environment Variables END:***********************************")

	if CLIENT_ID != "" {
		globalEnvVars["CLIENT_ID"] = CLIENT_ID
		globalEnvVars["CLIENT_SECRET"] = CLIENT_SECRET
		globalEnvVars["SUBSCRIPTION_ID"] = SUBSCRIPTION_ID
		globalEnvVars["TENANT_ID"] = TENANT_ID
	}

	return globalEnvVars, nil
}

func TestTerraform_azure_subscription(t *testing.T) {
	t.Parallel()
	envVars, err := setTerraformVariables()
	if err != nil {
		fmt.Printf("Error setting Terraform variables: %v\n", err)
		return
	}
	fmt.Println("Environment Variables:")
	for key, value := range envVars {
		fmt.Printf("%s: %s\n", key, value)
	}

	subscriptionID := envVars["SUBSCRIPTION_ID"]
	if subscriptionID == "" {
		fmt.Println("Azure subscription ID is not set")
		return
	}
	terraformOptions := terraform.WithDefaultRetryableErrors(t, &terraform.Options{
		TerraformDir: "../../terraform/module",
		Vars: map[string]interface{}{
			"subscription_name":          subscription_name,
			"workload":                   workload,
			"enrollment_account_name":    enrollment_account_name,
			"managementgroupassociation": managementgroupassociation,
			"billing_account_name":       billing_account_name,
		},
		EnvVars:       globalEnvVars,
		BackendConfig: globalBackendConf,
		NoColor:       true,

		Reconfigure: true,
	})

	defer terraform.Destroy(t, terraformOptions)

	terraform.InitAndApply(t, terraformOptions)
	expectedsubscription_id := terraform.Output(t, terraformOptions, "subscription_id")
	expectedsubscription_name := terraform.Output(t, terraformOptions, "subscription_name")
	expectedazuerm_subscription_tenant_id := terraform.Output(t, terraformOptions, "azuerm_subscription_tenant_id")

	fmt.Println("PRINTING THE RESOURCE PROPERTIES FROM OUTPUT FILE......................................")

	fmt.Printf("subscription_id : %s\n", expectedsubscription_id)
	fmt.Printf("subscription_name : %s\n", expectedsubscription_name)
	fmt.Printf("azuerm_subscription_tenant_id : %s\n", expectedazuerm_subscription_tenant_id)
	fmt.Println("PRINTING THE RESOURCE PROPERTIES FROM OUTPUT FILE HAS BEEN ENDED........................")
	fmt.Println("Token subscription id........................")

	fmt.Printf("Subscription ID: %s\n", subscriptionID)
	fmt.Println("Token subscription id........................")

	accessToken, err := getAccessToken(subscriptionID)
	if err != nil {
		fmt.Printf("Failed to get access token: %s\n", err.Error())
		return
	}
	url := fmt.Sprintf("https://management.azure.com/subscriptions/%s?api-version=%s", expectedsubscription_id, apiVersion)
	fmt.Println("Here the URL for the subscription Module Rest API:", url)

	subscriptionJSON, err := getSubscriptionDetails(url, accessToken)
	if err != nil {
		fmt.Printf("Failed to get subscription details: %s\n", err.Error())
		return
	}

	actual_data, err := printSubscriptionDetails(subscriptionJSON)
	if err != nil {
		log.Fatalf("failed to obtain a tha values from the RESTAPI: %v", err)
	}

	fmt.Println("\nSubscription Details from REST API - opened :-----------------")
	fmt.Printf("Display Name: %s\n", actual_data.DisplayName)
	fmt.Printf("Tenant ID: %s\n", actual_data.TenantID)
	fmt.Println("\nSubscription Details from REST API - closed :-----------------")

	// Test cases : -
	fmt.Println("Test cases are  running........")
	t.Run("Subscription_Name has been matched..", func(t *testing.T) {
		assert.Equal(t, expectedsubscription_name, actual_data.DisplayName)
	})

	t.Run("Tenant_Id has been matched..", func(t *testing.T) {
		assert.Equal(t, expectedazuerm_subscription_tenant_id, actual_data.TenantID)
	})

}

func getAccessToken(subscriptionID string) (string, error) {
	cmd := exec.Command("az", "account", "get-access-token", "--query", "accessToken", "--output", "tsv", "--subscription", subscriptionID)

	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(output)), nil
}

func getSubscriptionDetails(url, accessToken string) ([]byte, error) {
	client := &http.Client{}
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}

func printSubscriptionDetails(subscriptionJSON []byte) (Subscription, error) {

	fmt.Println("#############################---Complete JSON Response from the RESTAPI---#######################################")

	fmt.Println(string(subscriptionJSON))

	fmt.Println("#############################---Complete JSON Response from the RESTAPI---#######################################")

	var subscription Subscription
	err := json.Unmarshal(subscriptionJSON, &subscription)
	if err != nil {
		fmt.Printf("Failed to unmarshal JSON response: %s\n", err.Error())
		return Subscription{}, err
	}
	return subscription, err

}
