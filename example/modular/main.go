// Package main demonstrates focused modular usage of PocketBase sub-packages.
//
// This example shows how to import and use specific sub-packages independently
// for more focused functionality and better code organization.
//
// Run with: make serve (in another terminal) && go run example/modular_usage.go
//
// +build ignore

package main

import (
	"log"

	"github.com/Forty2Co/pocketbase"
	"github.com/Forty2Co/pocketbase/collections"
	"github.com/Forty2Co/pocketbase/realtime"
)

type User struct {
	ID       string `json:"id"`
	Email    string `json:"email"`
	Name     string `json:"name"`
	Created  string `json:"created"`
	Updated  string `json:"updated"`
}

func main() {
	log.Println("=== Modular Sub-Package Usage Examples ===")
	
	// Create base client for shared resources
	client := pocketbase.NewClient("http://localhost:8090")
	
	// Example 1: Collections-focused usage
	demonstrateCollectionsPackage(client)
	
	// Example 2: Admin-focused usage (requires admin auth)
	demonstrateAdminPackage(client)
	
	// Example 3: Realtime-focused usage
	demonstrateRealtimePackage(client)
}

// demonstrateCollectionsPackage shows focused collections usage
func demonstrateCollectionsPackage(client *pocketbase.Client) {
	log.Println("\n--- Collections Package Usage ---")
	
	// Create type-safe collection wrapper
	users := collections.CollectionSet[User](client, "users")
	
	log.Println("1. Listing users with type safety...")
	response, err := users.List(pocketbase.ParamsList{
		Size: 5,
		Sort: "-created",
	})
	if err != nil {
		log.Printf("   Error listing users: %v", err)
		return
	}
	
	log.Printf("   Found %d users\n", len(response.Items))
	for i, user := range response.Items {
		log.Printf("   User %d: %s (%s)\n", i+1, user.Name, user.Email)
	}
	
	// Example of authentication methods (if collection supports it)
	log.Println("2. Collection authentication methods...")
	// In a real scenario with user collection:
	// methods, err := users.AuthMethods()
	// if err == nil {
	//     log.Printf("   Email/password auth: %v\n", methods.EmailPassword)
	//     log.Printf("   OAuth providers: %d available\n", len(methods.AuthProviders))
	// }
	log.Println("   Authentication methods available for user collections")
}

// demonstrateAdminPackage shows focused admin operations
func demonstrateAdminPackage(client *pocketbase.Client) {
	log.Println("\n--- Admin Package Usage ---")
	log.Println("Note: Admin operations require admin authentication")
	
	log.Println("1. Backup operations...")
	// In a real scenario with admin auth:
	// backup := client.Backup()
	// backups, err := backup.FullList()
	// if err == nil {
	//     log.Printf("   Available backups: %d\n", len(backups))
	// }
	
	// Example backup creation (requires admin auth)
	// backupName := "example_backup_" + time.Now().Format("20060102_150405") + ".zip"
	// err := backup.Create(backupName)
	// if err == nil {
	//     log.Printf("   Created backup: %s\n", backupName)
	// }
	
	log.Println("2. File operations...")
	// Example file token generation (requires admin auth)
	// files := client.Files()
	// token, err := files.GetToken()
	// if err == nil {
	//     log.Printf("   File access token generated\n")
	// }
	
	log.Println("   Admin operations available: backup management, file tokens")
}

// demonstrateRealtimePackage shows focused real-time functionality
func demonstrateRealtimePackage(client *pocketbase.Client) {
	log.Println("\n--- Realtime Package Usage ---")
	
	log.Println("1. Setting up real-time subscription...")
	
	// Example subscription setup (in real usage, this would be in a goroutine)
	// users := collections.CollectionSet[User](client, "users")
	// stream, err := users.Subscribe()
	// if err != nil {
	//     log.Printf("   Error creating subscription: %v", err)
	//     return
	// }
	// defer stream.Unsubscribe()
	
	log.Printf("   Collection '%s' ready for real-time subscriptions\n", "users")
	// 
	// // Wait for connection
	// <-stream.Ready()
	// log.Println("   Connected to real-time stream")
	// 
	// // Handle events (in real usage, this would run in a loop)
	// go func() {
	//     for event := range stream.Events() {
	//         log.Printf("   Event: %s on record %s\n", event.Action, event.Record.ID)
	//     }
	// }()
	
	log.Println("2. Subscription options...")
	// Example with custom options
	opts := realtime.SubscribeOptions{}
	_ = opts // Use opts in real subscription
	
	log.Println("   Real-time subscriptions support custom headers and options")
	log.Println("   Events include: create, update, delete actions")
	log.Println("   Automatic reconnection and error handling included")
}