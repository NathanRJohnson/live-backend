package item

import (
	"context"
	"log"
	"strconv"

	"cloud.google.com/go/firestore"
	"github.com/NathanRJohnson/live-backend/wtfridge/model"
	"google.golang.org/api/iterator"
)

type FirebaseRepo struct {
	Client *firestore.Client
}

func (r *FirebaseRepo) Insert(ctx context.Context, collection string, item model.Item) error {
	_, err := r.Client.Collection(collection).Doc(strconv.Itoa(item.GetID())).Set(ctx, item)
	if err != nil {
		log.Fatalf("Failed adding item: %v", err)
	}

	return nil
}

func (r *FirebaseRepo) FetchAll(ctx context.Context, collection string) ([]interface{}, error) {
	var items []interface{}
	iter := r.Client.Collection(collection).Documents(ctx)

	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}

		// gets empty item
		item := getItemSchemaByCollection(collection)
		if item == nil {
			log.Fatal("error getting item schema for collection:", collection)
		}
		// fills item with document data
		err = doc.DataTo(&item)
		if err != nil {
			log.Fatalf("error unmarshalling document to item representation: %v", err)
		}

		items = append(items, item)
	}
	return items, nil
}

func (r *FirebaseRepo) DeleteByID(ctx context.Context, collection string, item model.Item) error {
	_, err := r.Client.Collection(collection).Doc(strconv.Itoa(item.GetID())).Delete(ctx)
	if err != nil {
		log.Fatalf("unable to delete item: %v", err)
		return err
	}
	return nil
}

// TODO: update this to take any path and any value
func (r *FirebaseRepo) ToggleActiveByID(ctx context.Context, collection string, id int) error {
	ref := r.Client.Collection(collection).Doc(strconv.Itoa(id))
	err := r.Client.RunTransaction(ctx, func(ctx context.Context, tx *firestore.Transaction) error {
		doc, err := tx.Get(ref)
		if err != nil {
			log.Printf("unable to get %d from %s", id, collection)
			return err
		}

		data, err := doc.DataAt("IsActive")
		if err != nil {
			log.Printf("unable to read is_active field: %v", err)
			return err
		}

		is_active, ok := data.(bool)
		if !ok {
			log.Printf("unable to convert data to bool")
			is_active = false
		}

		updates := []firestore.Update{
			{Path: "IsActive", Value: !is_active},
		}

		return tx.Update(ref, updates)
	})

	if err != nil {
		log.Printf("unable to toggle activitiy: %v", err)
	}

	return err
}

func getItemSchemaByCollection(collection string) interface{} {
	switch collection {
	case "fridge":
		return &model.FridgeItem{}
	case "grocery":
		return &model.GroceryItem{}
	default:
		return nil
	}
}
