package item

import (
	"context"
	"errors"
	"fmt"
	"log"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"cloud.google.com/go/firestore"
	"cloud.google.com/go/firestore/apiv1/firestorepb"
	"github.com/NathanRJohnson/live-backend/wtfridge/model"
	"google.golang.org/api/iterator"
)

type FirebaseRepo struct {
	Client *firestore.Client
}

func (r *FirebaseRepo) GetDocRef(collection *firestore.CollectionRef, doc string) *firestore.DocumentRef {
	return collection.Doc(doc)
}

func (r *FirebaseRepo) GetCollectionRef(collection string, doc *firestore.DocumentRef) *firestore.CollectionRef {
	var collectionRef *firestore.CollectionRef = nil
	if doc == nil {
		collectionRef = r.Client.Collection(collection)
	} else {
		collectionRef = doc.Collection(collection)
	}
	return collectionRef
}

// func (r *FirebaseRepo) getUserDocRef(ctx context.Context, user model.User) (*firestore.DocumentRef, error) {
// 	userDocRef := r.Client.Collection("USER").Doc(user.Username)
// 	if !docExists(ctx, userDocRef) {
// 		return nil, errors.New("could not get user doc")
// 	}
// 	return userDocRef, nil
// }

func (r *FirebaseRepo) DocExists(ctx context.Context, doc *firestore.DocumentRef) (bool, error) {
	snapshot, err := doc.Get(ctx)
	if err != nil {
		return false, err
	}
	return snapshot.Exists(), nil
}

func (r *FirebaseRepo) Insert(ctx context.Context, collection interface{}, data map[string]interface{}) error {
	var collectionRef *firestore.CollectionRef
	if c, ok := collection.(*firestore.CollectionRef); !ok {
		return errors.New("must pass interface of type firestoreCollectionRef into Insert")
	} else {
		collectionRef = c
	}

	_, _, err := collectionRef.Add(ctx, data)
	if err != nil {
		log.Printf("Failed adding item: %v", err)
	}

	return nil
}

func (r *FirebaseRepo) FetchAll(ctx context.Context, collection interface{}) ([]interface{}, error) {
	var collectionRef *firestore.CollectionRef
	if c, ok := collection.(*firestore.CollectionRef); !ok {
		return nil, errors.New("must pass interface of type firestoreCollectionRef into FetchAll")
	} else {
		collectionRef = c
	}

	var items []interface{}
	iter := collectionRef.Documents(ctx)

	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}

		// gets empty item
		schema := getItemSchemaFromCollection(collectionRef.Path)
		if schema == nil {
			log.Fatal("error getting item schema for collection:", collection)
		}
		// fills item with document data
		err = doc.DataTo(&schema)
		if err != nil {
			log.Fatalf("error unmarshalling document to item representation: %v", err)
		}

		items = append(items, schema)
	}
	return items, nil
}

func (r *FirebaseRepo) DeleteByID(ctx context.Context, collection interface{}, id int) error {
	var collectionRef *firestore.CollectionRef
	if c, ok := collection.(*firestore.CollectionRef); !ok {
		return errors.New("must pass interface of type firestoreCollectionRef into DeleteByID")
	} else {
		collectionRef = c
	}

	docs, err := collectionRef.Where("ItemID", "==", id).Documents(ctx).GetAll()
	if err != nil {
		return err
	} else if len(docs) == 0 {
		return errors.New("failed to delete document: not found")
	} else if len(docs) > 1 {
		return errors.New("failed to delete document: multiple documents found with matching IDs")
	}

	doc := docs[0]

	// shift indicies up to account for the deleted item
	if strings.ToUpper(collectionRef.ID) == "GROCERY" {
		data, err := doc.DataAt("Index")
		if err != nil {
			return err
		}

		removed_index, _ := data.(int64)

		err = r.shiftIndicies(ctx, collectionRef, -1, int(removed_index), 100)
		if err != nil {
			return err
		}
	}

	_, err = doc.Ref.Delete(ctx)
	if err != nil {
		log.Printf("unable to delete item: %v", err)
		return err
	}

	return nil
}

// TODO: update this to take any path and any value
func (r *FirebaseRepo) ToggleActiveByID(ctx context.Context, collection interface{}, id int) error {
	var collectionRef *firestore.CollectionRef
	if c, ok := collection.(*firestore.CollectionRef); !ok {
		return errors.New("must pass interface of type firestoreCollectionRef into DeleteByID")
	} else {
		collectionRef = c
	}

	docs, err := collectionRef.Where("ItemID", "==", id).Documents(ctx).GetAll()
	if err != nil {
		return err
	} else if len(docs) == 0 {
		return errors.New("failed to update document: not found")
	} else if len(docs) > 1 {
		return errors.New("failed to update document: multiple documents found with matching IDs")
	}

	doc := docs[0]
	err = r.Client.RunTransaction(ctx, func(ctx context.Context, tx *firestore.Transaction) error {
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

		return tx.Update(doc.Ref, updates)
	})

	if err != nil {
		log.Printf("unable to toggle activitiy: %v", err)
	}

	return err
}

func (r *FirebaseRepo) UpdateItemByID(ctx context.Context, collection interface{}, id int, itemValues map[string]interface{}) error {
	var collectionRef *firestore.CollectionRef
	if c, ok := collection.(*firestore.CollectionRef); !ok {
		return errors.New("must pass interface of type firestoreCollectionRef into FetchAll")
	} else {
		collectionRef = c
	}

	docs, err := collectionRef.Where("ItemID", "==", id).Documents(ctx).GetAll()
	if err != nil {
		return err
	} else if len(docs) == 0 {
		return errors.New("failed to update document: not found")
	} else if len(docs) > 1 {
		return errors.New("failed to update document: multiple documents found with matching IDs")
	}

	doc := docs[0]
	err = r.Client.RunTransaction(ctx, func(ctx context.Context, tx *firestore.Transaction) error {

		var updates []firestore.Update

		for path, newValue := range itemValues {
			updates = append(updates, firestore.Update{Path: path, Value: newValue})
		}

		return tx.Update(doc.Ref, updates)
	})

	if err != nil {
		log.Printf("unable to update item: %v", err)
	}

	return err
}

func (r *FirebaseRepo) MoveToFridge(ctx context.Context, source interface{}, dest interface{}) error {

	var sourceRef *firestore.CollectionRef
	if c, ok := source.(*firestore.CollectionRef); !ok {
		return errors.New("must pass interface of type firestoreCollectionRef into FetchAll")
	} else {
		sourceRef = c
	}

	var destRef *firestore.CollectionRef
	if c, ok := dest.(*firestore.CollectionRef); !ok {
		return errors.New("must pass interface of type firestoreCollectionRef into FetchAll")
	} else {
		destRef = c
	}

	active_doc_refs := sourceRef.Where("IsActive", "==", true).Documents(ctx)

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

			destItemRef := firestore.DocumentRef{
				Parent: destRef,
				Path:   filepath.Join(destRef.Path, strconv.Itoa(grocery_item.ItemID)),
				ID:     strconv.Itoa(grocery_item.ItemID),
			}

			err = tx.Create(&destItemRef, fridge_item)
			if err != nil {
				log.Printf("unable to create fridge item %s: %v", fridge_item.Name, err)
				return err
			}
		}
	})

	if err != nil {
		return err
	}

	return err
}

func (r *FirebaseRepo) remapIndiciesFromRemoved(ctx context.Context, removed_indicies map[int]bool) error {
	if len(removed_indicies) == 0 {
		return nil
	}

	// temp
	groceryCollection := r.GetCollectionRef("grocery", nil)

	num_items, err := r.countNumDocs(ctx, groceryCollection)
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

func (r *FirebaseRepo) countNumDocs(ctx context.Context, collection *firestore.CollectionRef) (int, error) {
	query := collection.NewAggregationQuery().WithCount("all")
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

func (r *FirebaseRepo) RearrageItems(ctx context.Context, collection interface{}, old_index int64, new_index int64) error {
	var collectionRef *firestore.CollectionRef
	if c, ok := collection.(*firestore.CollectionRef); !ok {
		return errors.New("must pass interface of type firestoreCollectionRef into FetchAll")
	} else {
		collectionRef = c
	}

	max_index, err := r.countNumDocs(ctx, collectionRef)
	if err != nil {
		return err
	}

	if new_index > int64(max_index) || old_index > int64(max_index) {
		return errors.New("indicies exceed max index")
	}

	doc, err := collectionRef.Where("Index", "==", old_index).Documents(ctx).Next()
	if err != nil {
		return fmt.Errorf("could not find document with index %d", old_index)
	}

	if new_index < old_index { // moved up
		err = r.shiftIndicies(ctx, collectionRef, 1, int(new_index), int(old_index)-1)
	} else { // moved down
		err = r.shiftIndicies(ctx, collectionRef, -1, int(old_index)+1, int(new_index))
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

func (r *FirebaseRepo) shiftIndicies(ctx context.Context, collection *firestore.CollectionRef, amount int64, start_index int, end_index int) error {
	err := r.Client.RunTransaction(ctx, func(ctx context.Context, tx *firestore.Transaction) error {
		docs := collection.Where("Index", ">=", start_index).Where("Index", "<=", end_index)
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

func getItemSchemaFromCollection(collectionPath string) interface{} {
	splits := strings.Split(collectionPath, "/")
	collection := strings.ToUpper(splits[len(splits)-1])

	switch collection {
	case "FRIDGE":
		return &model.FridgeItem{}
	case "GROCERY":
		return &model.GroceryItem{}
	default:
		return nil
	}
}

func (r *FirebaseRepo) CreateUser(ctx context.Context, user model.User) error {
	_, err := r.Client.Collection("USER").Doc(user.Username).Set(ctx, user)
	if err != nil {
		log.Printf("Failed to create user: %v", err)
	}
	return err
}
