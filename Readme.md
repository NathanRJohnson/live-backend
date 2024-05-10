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


## Pulling containers to EC2 from private GHCR
On your local machine
1. Login into docker using the following this guide: https://docs.github.com/en/packages/working-with-a-github-packages-registry/working-with-the-container-registry
2. add the image field to the service of the image you want to build
3. build the image: `docker compose build --no-cache <LOCAL_IMAGE_NAME>`
4. tag the image: `docker tag <LOCAL_IMAGE_NAME> ghcr.io/<GHCR_UNAME>/<IMAGE_NAME>:<TAG>`
5. push the image: `docker push ghcr.io/<GHCR_UNAME>/<IMAGE_NAME>:<TAG>`

Then, SSH into the EC2 intance
1. If docker is not already installed / configured, you will have to do so. Docker gets funny about authentication when trying to run using sudo, so the following will remove the need for sudo.
  * Install docker: https://docs.docker.com/engine/install/ubuntu/
  * Start docker: `sudo service docker start`
  * Add user to docker group: `sudo usermod -aG docker $(whoami)`
  * Relog to update permissions: `sudo su $(whoami)`
    * If you do not know / have not setup a password: https://stackoverflow.com/questions/51667876/ec2-ubuntu-14-default-password 
2. Login into docker using the following this guide: https://docs.github.com/en/packages/working-with-a-github-packages-registry/working-with-the-container-registry
3. Pull the image: `docker pull ghcr.io/<GHCR_UNAME>/<IMAGE_NAME>:<TAG>`