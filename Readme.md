### Project L

## Deployment

# Prerequisits
This app makes use of Firestore, and thus requires auth to access the database via a firestore service account. Auth credentials must be provided at `secrets/firebase-serviceKey.json`, which will be stored as a secret in the appropriate service(s).

# Starting
To start the project in your local environemnt:
1. Start the docker daemon `sudo service docker start`
2. Start the containers `docker compose up -d`
3. The project will be available at `http://localhost`

# Stopping
To stop the project in your local environment:
1. Stop the containers `docker compose down`