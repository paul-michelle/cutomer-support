### Dev
Prepare a database instance:
```
docker run --rm -d -p 127.0.0.1:5433:5432 -e POSTGRES_USER=... -e POSTGRES_PASSWORD=... -e POSTGRES_DB=... postgres:13-alpine
```
Add a .env file to the project root, specifying DB_HOST=..., DB_PORT=..., DB_USER=..., DB_PASSWORD=..., DB_NAME=....
The default server address will be "localhost:8089". Provide SERVER_HOST=... and SERVER_PORT=... values to change it. 
Mind that variables from the docker run command should correspond to what is being provided in the .env file.
From the project root fire:
```
go run .
```
Make a ping request to http://localhost:8089/time to get the current time (open in the browser for quick check).

### Users
Users are registere via an API call (POST /users) with the following info in request payload: unique email, at least 8 chars password, and username. By additionally passing true to isStaff boolean field, the user to be registered is claimed to have a stuff status, in which case a bearer token in auth headers expected.

For safety reasons, any other fields added to the request body, incl. isSuperuser, will be ignored.

All users are considered to have common status, unless is_staff or is_superuser is true in the corresponding 
field of the users table.

### Authorization
To receive a JWT, a post request to /login endpoint expected with email and password specified.