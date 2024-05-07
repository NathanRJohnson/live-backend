### What the fridge
A personal food management app with the goal of reducing food waste and promoting better eating habits.

## Deployment

# Prerequisits
This app makes use of Firestore, and thus requires auth to access the database via a firestore service account. Auth credentials must be provided at `secrets/firebase-serviceKey.json` *one level above the root of this project*.

This app requires go v1.22.x

# Starting
1. `go run main.go`
2. Application will be available at `http://localhost:3000/fridge` 

# Stopping
The application implments graceful shut down, so using `ctrl+C` is perfectly fine.