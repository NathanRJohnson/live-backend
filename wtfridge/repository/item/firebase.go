package item

import (
	"context"
	"log"

	"cloud.google.com/go/firestore"
	"github.com/NathanRJohnson/live-backend/wtfridge/model"
)

type FirebaseRepo struct {
	Client *firestore.Client
}

func (r *FirebaseRepo) Insert(ctx context.Context, item model.Item) error {
	_, _, err := r.Client.Collection("fridge").Add(ctx, item)
	if err != nil {
		log.Fatalf("Failed adding item: %v", err)
	}

	return nil
}
