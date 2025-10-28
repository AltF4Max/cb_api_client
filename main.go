package main

import (
	"cb_api_client/internal/client"
	"context"
	"fmt"
	"log"
)

func main() {
	config := &client.CleverbridgeConfig{
		ClientID:     "your_cleverbridge_client_id",
		ClientSecret: "your_cleverbridge_client_secret",
		BaseURL:      "https://rest.cleverbridge.com",
		Debug:        true,
	}

	cbClient := client.NewBaseClient(config)

	ctx := context.Background()

	//fmt.Println("1. 📦 Getting subscription by ID...")
	subscription, err := cbClient.GetSubscription(ctx, "S18577447", "false")
	if err != nil {
		log.Printf("⚠️ Error getting subscription: %v", err)
	} else {
		fmt.Printf("✅ Subscription: ID=%s, Status=%s, Plan=%s\n",
			subscription.ID, subscription.Status, subscription.Plan)
	}

	//fmt.Println("\n2. 🛒 Getting subscriptions by purchase...")
	purchaseSubscriptions, err := cbClient.GetSubscriptionsByPurchase(ctx, "P123456789")
	if err != nil {
		log.Printf("⚠️ Error getting subscriptions by purchase: %v", err)
	} else {
		fmt.Printf("✅ Found %d subscriptions for purchase\n", len(purchaseSubscriptions))
		for i, sub := range purchaseSubscriptions {
			fmt.Printf("   %d. %s - %s\n", i+1, sub.ID, sub.Status)
		}
	}

	//fmt.Println("\n3. 👤 Getting subscriptions for customer...")
	customerSubscriptions, err := cbClient.GetSubscriptionsForCustomer(ctx, "CUST12345")
	if err != nil {
		log.Printf("⚠️ Error getting subscriptions for customer: %v", err)
	} else {
		fmt.Printf("✅ Found %d subscriptions for customer\n", len(customerSubscriptions))
		for i, sub := range customerSubscriptions {
			fmt.Printf("   %d. %s - %s - %s\n", i+1, sub.ID, sub.Status, sub.Plan)
		}
	}
}
