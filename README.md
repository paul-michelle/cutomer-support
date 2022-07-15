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
Users are registered via an API call (POST /users) with the following info in request payload: unique email, at least 8 chars password, and username. 
```
POST /users
{
    "email": "valid@format.here",
    "password": "atLeastEightChars",
    "username": "whoami",
}
```
By additionally passing true to isStaff boolean field, the user to be registered is claimed to have a stuff status, in which case a bearer token in auth headers expected.
```
POST /users
Authorization: Bearer <Secret here>
{
    "email": "valid@format.here",
    "password": "atLeastEightChars",
    "username": "whoami",
    "isStaff": true
}
```
201 Created || 405 Method Not Allowed || 401 Unauthorized (when creating stuff member) || 400 Bad Request with error details.

#### Superuser considerations
For safety reasons, any other fields added to the request body, incl. isSuperuser, will be ignored.
Superusers are currently not to be created the way common users and staff are.
By design, admin/supersusers will be able to create and provide members of the stuff with the API key to be used in auth headers (currently, it's one single API key passed as an envvar and checked against).

Btw, all users are considered to have common status, unless is_staff or is_superuser is true in the corresponding 
field of the users table.

### Authorization
To receive a JWT, a post request to /login endpoint expected with email and password specified.
```
POST /users
{
    "email": "valid@format.here",
    "password": "atLeastEightChars",
}
```
200 OK || 405 Method Not Allowed || 404 Not Found || 400 Bad Request || 500 Internal Server Error (jwt string creation issue, user is asked to retry)

### Flow
1. Create a user: POST /users (no jwt)
2. Login to set jwt token as cookie: POST /login (no jwt)
3. 
