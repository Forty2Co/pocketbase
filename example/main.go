// Package main provides usage examples for the PocketBase Go SDK.
//
// This example demonstrates both unified client usage and modular sub-package usage
// for basic CRUD operations, real-time subscriptions, and admin operations.
//
// Run with: make serve (in another terminal) && go run example/main.go
package main

import (
	"errors"
	"log"
	"time"

	"github.com/Forty2Co/pocketbase"
	"github.com/Forty2Co/pocketbase/collections"
	"github.com/mitchellh/mapstructure"
)

type Post struct {
	Field   string `json:"field"`
	ID      string `json:"id"`
	Created string `json:"created"`
	Updated string `json:"updated"`
}

func main() {
	// REMEMBER to start the Pocketbase before running this example with `make serve` command

	log.Println("=== PocketBase Go SDK Examples ===")
	log.Println("Demonstrating both unified client and modular sub-package usage")
	
	// Run unified client examples
	if err := runUnifiedClientExamples(); err != nil {
		log.Printf("Unified client examples failed: %v", err)
	}
	
	// Run modular sub-package examples
	if err := runModularExamples(); err != nil {
		log.Printf("Modular examples failed: %v", err)
	}
}

// runUnifiedClientExamples demonstrates the traditional unified client approach
func runUnifiedClientExamples() error {
	log.Println("\n--- Unified Client Usage (Backward Compatible) ---")
	
	var errs error
	client := pocketbase.NewClient("http://localhost:8090")
	// Configuration options remain the same:
	// pocketbase.WithAdminEmailPassword("admin@admin.com", "admin@admin.com")
	// pocketbase.WithUserEmailPassword("user@user.com", "user@user.com")
	// pocketbase.WithUserToken(token)
	// pocketbase.WithAdminToken(token)
	// pocketbase.WithRestDebug() // Enable HTTP request/response debugging
	// pocketbase.WithSseDebug()  // Enable Server-Sent Events debugging

	log.Println("1. Listing records using unified client...")
	response, err := client.List("posts_public", pocketbase.ParamsList{
		Size:    2,
		Page:    1,
		Sort:    "-created",
		Filters: "field~'test'",
	})
	errs = errors.Join(errs, err)

	log.Printf("   Found %d items, %d total pages\n", len(response.Items), response.TotalPages)
	for i, item := range response.Items {
		var post Post
		err := mapstructure.Decode(item, &post)
		errs = errors.Join(errs, err)
		log.Printf("   Item %d: %s (ID: %s)\n", i+1, post.Field, post.ID)
	}

	log.Println("2. Creating record using unified client...")
	newPost, err := client.Create("posts_public", Post{
		Field: "unified_test_" + time.Now().Format("15:04:05"),
	})
	errs = errors.Join(errs, err)
	if err == nil {
		log.Printf("   Created post with ID: %s\n", newPost.ID)
		
		// Clean up
		err = client.Delete("posts_public", newPost.ID)
		errs = errors.Join(errs, err)
		if err == nil {
			log.Printf("   Cleaned up post: %s\n", newPost.ID)
		}
	}

	log.Println("3. Using type-safe collections (unified approach)...")
	collection := pocketbase.CollectionSet[Post](client, "posts_public")
	typedResponse, err := collection.List(pocketbase.ParamsList{Size: 1})
	errs = errors.Join(errs, err)
	if err == nil {
		log.Printf("   Type-safe collection found %d items\n", len(typedResponse.Items))
		for _, post := range typedResponse.Items {
			log.Printf("   Typed post: %s (ID: %s)\n", post.Field, post.ID)
		}
	}

	return errs
}

// runModularExamples demonstrates the new modular sub-package approach
func runModularExamples() error {
	log.Println("\n--- Modular Sub-Package Usage (New Approach) ---")
	
	var errs error
	
	// Create base client for shared resources (HTTP client, auth, config)
	client := pocketbase.NewClient("http://localhost:8090")
	
	log.Println("1. Using collections sub-package directly...")
	// Access collections functionality through sub-package
	collection := collections.CollectionSet[Post](client, "posts_public")
	
	response, err := collection.List(pocketbase.ParamsList{
		Size: 2,
		Sort: "-created",
	})
	errs = errors.Join(errs, err)
	if err == nil {
		log.Printf("   Collections sub-package found %d items\n", len(response.Items))
		for i, post := range response.Items {
			log.Printf("   Item %d: %s (Created: %s)\n", i+1, post.Field, post.Created)
		}
	}

	log.Println("2. Creating record using collections sub-package...")
	newPost := Post{
		Field: "modular_test_" + time.Now().Format("15:04:05"),
	}
	created, err := collection.Create(newPost)
	errs = errors.Join(errs, err)
	if err == nil {
		log.Printf("   Created via sub-package: %s (ID: %s)\n", created.Field, created.ID)
		
		// Clean up
		err = collection.Delete(created.ID)
		errs = errors.Join(errs, err)
		if err == nil {
			log.Printf("   Cleaned up via sub-package: %s\n", created.ID)
		}
	}

	log.Println("3. Using realtime sub-package for subscriptions...")
	// Demonstrate realtime functionality (would need actual subscription in real use)
	log.Println("   Realtime client available for SSE subscriptions")
	// In a real scenario, you would:
	// stream, err := collection.Subscribe()
	// if err == nil {
	//     defer stream.Unsubscribe()
	//     // Handle events from stream.Events()
	// }

	log.Println("4. Using admin sub-package for operations...")
	// Access admin functionality through sub-package
	log.Println("   Admin client available for backup and file operations")
	// In a real scenario with admin auth, you would:
	// backup := client.Backup()
	// files := client.Files()
	// backups, err := backup.FullList()

	log.Println("\n--- Comparison Summary ---")
	log.Println("• Unified approach: All functionality through main client (backward compatible)")
	log.Println("• Modular approach: Import and use specific sub-packages as needed")
	log.Println("• Both approaches share the same HTTP client, auth, and configuration")
	log.Println("• Choose based on your preference: unified simplicity vs modular organization")

	return errs
}
