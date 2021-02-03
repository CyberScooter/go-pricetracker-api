package main

import (
	"github.com/tomogoma/go-api-guard"
	"fmt"
)

func main(){

	db := &DBMock{} //implements KeyStore interface

	// mocking key generation to demonstrate resulting API key
	keyGen := &KeyGenMock{ExpSRBs: []byte("an-api-key")}
	
	g, _ := api.NewGuard(
		db,
		api.WithKeyGenerator(keyGen), // This is optional
	)
	
	// Generate API key
	APIKey, _ := g.NewAPIKey("my-unique-user-id")
	
	fmt.Println(string(APIKey.Value()))
	
	// Validate API Key
	userID, _ := g.APIKeyValid(APIKey.Value())
	
	fmt.Println(userID)

}