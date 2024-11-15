package item

import (
	"context"
	"errors"
	"fmt"
	"log"
	"path/filepath"
	"strconv"
	"time"

	"cloud.google.com/go/firestore"
	"cloud.google.com/go/firestore/apiv1/firestorepb"
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

func (r *FirebaseRepo) DeleteByID(ctx context.Context, collection string, id int) error {
	ref := r.Client.Collection(collection).Doc(strconv.Itoa(id))

	// shift indicies up to account for the deleted item
	if collection == "grocery" {
		doc, err := ref.Get(ctx)
		if err != nil {
			return err
		}

		data, err := doc.DataAt("Index")
		if err != nil {
			return err
		}

		removed_index, _ := data.(int64)

		err = r.shiftIndicies(ctx, collection, -1, int(removed_index), 100)
		if err != nil {
			return err
		}
	}

	_, err := ref.Delete(ctx)
	if err != nil {
		log.Printf("unable to delete item: %v", err)
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
			return errors.New("unable to convert data to bool")
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

func (r *FirebaseRepo) UpdateItemByID(ctx context.Context, collection string, item_id int, item_values map[string]interface{}) error {
	ref := r.Client.Collection(collection).Doc(strconv.Itoa(item_id))
	err := r.Client.RunTransaction(ctx, func(ctx context.Context, tx *firestore.Transaction) error {

		var updates []firestore.Update

		for path, new_value := range item_values {
			updates = append(updates, firestore.Update{Path: path, Value: new_value})
		}

		return tx.Update(ref, updates)
	})

	if err != nil {
		log.Printf("unable to update item: %v", err)
	}

	return err
}

func (r *FirebaseRepo) MoveToFridge(ctx context.Context) error {
	active_doc_refs := r.Client.Collection("grocery").Where("IsActive", "==", true).Documents(ctx)
	fridge_ref := r.Client.Collection("fridge")

	// using a map to act as a set
	removed_indicies := make(map[int]bool)

	err := r.Client.RunTransaction(ctx, func(ctx context.Context, tx *firestore.Transaction) error {
		for {
			doc, err := active_doc_refs.Next()
			if err == iterator.Done {
				return r.remapIndiciesFromRemoved(ctx, removed_indicies)
			} else if err != nil {
				log.Printf("could not iterate through docs: %v", err)
				return err
			}

			var grocery_item model.GroceryItem
			err = doc.DataTo(&grocery_item)
			if err != nil {
				log.Printf("unable to marshal data to grocery schema: %v", err)
				return err
			}

			removed_indicies[grocery_item.Index] = true

			err = tx.Delete(doc.Ref)
			if err != nil {
				log.Printf("unable to delete document %s: %v", doc.Ref.ID, err)
				return err
			}

			now := time.Now().UTC()
			fridge_item := model.FridgeItem{
				ItemID:    grocery_item.ItemID,
				Name:      grocery_item.Name,
				DateAdded: &now,
			}

			fridge_item_ref := firestore.DocumentRef{
				Parent: fridge_ref,
				Path:   filepath.Join(fridge_ref.Path, strconv.Itoa(grocery_item.ItemID)),
				ID:     strconv.Itoa(grocery_item.ItemID),
			}

			err = tx.Create(&fridge_item_ref, fridge_item)
			if err != nil {
				log.Printf("unable to create fridge item %s: %v", fridge_item.Name, err)
				return err
			}
		}

	})

	if err != nil {
		return err
	}

	log.Println(removed_indicies)

	return err
}

func (r *FirebaseRepo) remapIndiciesFromRemoved(ctx context.Context, removed_indicies map[int]bool) error {
	if len(removed_indicies) == 0 {
		return nil
	}

	num_items, err := r.countNumDocs(ctx, "grocery")
	if err != nil {
		return err
	}
	new_index_map := computeNewIndexMap(removed_indicies, num_items)

	iter := r.Client.Collection("grocery").Documents(ctx)

	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return err
		}

		data, err := doc.DataAt("Index")
		if err != nil {
			log.Printf("unable to read is_active field: %v", err)
			return err
		}

		old_index, ok := data.(int64)
		if !ok {
			return errors.New("unable to convert data to int")
		}
		log.Printf("hello: %d", int(old_index))
		new_index, ok := new_index_map[int(old_index)]
		if ok {
			log.Printf("updating with %d to %d", old_index, new_index)
			_, err = doc.Ref.Update(ctx, []firestore.Update{{Path: "Index", Value: new_index}})
			if err != nil {
				return err
			}
		}
	}
	return err
}

func (r *FirebaseRepo) countNumDocs(ctx context.Context, collection string) (int, error) {
	query := r.Client.Collection(collection).NewAggregationQuery().WithCount("all")
	results, err := query.Get(ctx)
	if err != nil {
		return 0, err
	}

	count, ok := results["all"]
	if !ok {
		return 0, errors.New("firestore: couldn't get alias for COUNT from results")
	}

	// makes sure the user doesn't send a request placing the item at an index beyond the size of the list
	num_docs := count.(*firestorepb.Value).GetIntegerValue()
	return int(num_docs), nil
}

func computeNewIndexMap(removed_indicies map[int]bool, num_items int) map[int]int {
	log.Print(removed_indicies)
	new_map := make(map[int]int)
	next_available_index := 1
	for i := 1; i <= num_items; i++ {
		_, ok := removed_indicies[i]
		if !ok {
			new_map[i] = next_available_index
			next_available_index++
		}
	}
	log.Print(new_map)
	return new_map
}

func (r *FirebaseRepo) RearrageItems(ctx context.Context, collection string, old_index int64, new_index int64) error {
	max_index, err := r.countNumDocs(ctx, collection)
	if err != nil {
		return err
	}

	if new_index > int64(max_index) || old_index > int64(max_index) {
		return errors.New("indicies exceed max index")
	}

	doc, err := r.Client.Collection(collection).Where("Index", "==", old_index).Documents(ctx).Next()
	if err != nil {
		return fmt.Errorf("could not find document with index %d", old_index)
	}

	if new_index < old_index { // moved up
		err = r.shiftIndicies(ctx, collection, 1, int(new_index), int(old_index)-1)
	} else { // moved down
		err = r.shiftIndicies(ctx, collection, -1, int(old_index)+1, int(new_index))
	}
	if err != nil {
		return err
	}

	_, err = doc.Ref.Update(ctx, []firestore.Update{{Path: "Index", Value: new_index}})
	if err != nil {
		return err
	}

	return err
}

func (r *FirebaseRepo) shiftIndicies(ctx context.Context, collection string, amount int64, start_index int, end_index int) error {
	err := r.Client.RunTransaction(ctx, func(ctx context.Context, tx *firestore.Transaction) error {
		docs := r.Client.Collection(collection).Where("Index", ">=", start_index).Where("Index", "<=", end_index)
		iter := docs.Documents(ctx)

		for {
			doc, err := iter.Next()
			if err == iterator.Done {
				return nil
			} else if err != nil {
				log.Printf("could not iterate through docs: %v", err)
				return err
			}

			data, err := doc.DataAt("Index")
			if err != nil {
				log.Printf("unable to read is_active field: %v", err)
				return err
			}

			err = tx.Update(
				doc.Ref,
				[]firestore.Update{{Path: "Index", Value: data.(int64) + amount}},
			)
			if err != nil {
				return err
			}
		}
	})
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
