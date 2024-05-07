package item

import (
	"context"
	"log"

	"cloud.google.com/go/firestore"
	"github.com/NathanRJohnson/live-backend/wtfridge/model"
	"google.golang.org/api/iterator"
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

func (r *FirebaseRepo) FetchAll(ctx context.Context) ([]model.Item, error) {
	var items []model.Item
	iter := r.Client.Collection("fridge").Documents(ctx)

	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}

		var item model.Item
		err = doc.DataTo(&item)
		if err != nil {
			log.Fatalf("error unmarhsalling document to item representation: %v", err)
		}
		items = append(items, item)
	}
	return items, nil
}
