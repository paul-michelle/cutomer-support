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

Is is only a staff member or a superuser who can get the list of all users (ordered DESC by the number of submitted tickets) via:
```
GET /users
```
The endpoint returns:

```
Status 200 OK
[
    {
        "id": 2,
        "created_at": "2022-07-16T07:26:15.592378Z",
        "username": "michelle",
        "email": "user@post.io",
        "is_staff": false,
        "tickets_count": 3
    },
    {
        "id": 1,
        "created_at": "2022-07-15T19:16:52.332915Z",
        "username": "michelle",
        "email": "anotheruser@post.io",
        "is_staff": false,
        "tickets_count": 2
    },
    {
        "id": 3,
        "created_at": "2022-07-16T07:30:25.527118Z",
        "username": "michelle",
        "email": "staffuser@post.io",
        "is_staff": true,
        "tickets_count": 0
    }
]
```
In case of failure: 405 Method Not Allowed || 401 Unauthorized (if not isStaff or isSuperuser) ||500 Internal Server Error (error reading from the database)

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

### Tickets
To create a ticket user needs to send a POST request to /tickets specifying topic (< 20 chars) and text:
```
POST /tickets
{
    "topic": "topic here",
    "text": "a long text with the issue here..."
}
```
Succesful response:
```
201 Created 
{
    "id": 10
}
```
Other possible responses: 405 Method Not Allowed || 401 Unauthorized (no or invalid jwt) || 400 Bad Request with error details || 500 Internal Server Error (error when writing to the db.)

To get all the tickets (all user's tickets for common user and a list of all existing tickets for staff member of superuser):
```
GET /tickets
```
Response:
```
200 OK
[
    {
        "id": 1,
        "created_at": "2022-07-16T07:12:30.676834Z",
        "updated_at": "2022-07-16T07:12:30.676834Z",
        "author": "postman33@mpost.io",
        "topic": "new instances",
        "status": "pending"
    },
    {
        "id": 2,
        "created_at": "2022-07-16T07:13:58.008489Z",
        "updated_at": "2022-07-16T07:13:58.008489Z",
        "author": "postman33@mpost.io",
        "topic": "new discount",
        "status": "pending"
    },
]
```
Other possible responses: 405 Method Not Allowed || 401 Unauthorized (no or invalid jwt) || || 500 Internal Server Error (err when reading from the database)

To get info on a particular ticket the users hits.
```
GET /tickets/{id}
```
In case the ticket exists and belongs to the user (or in case it is a  request from a staff member):
```
COMPLETE ME!!!
```

To change ticket status to "resolved" (which is the ONLY update that can be done to a ticket remotely):
```
PUT/PATCH /tickets/{id}
{
    "status": "resolved"
}
```
### Messaging
When the ticket is first registered, its text goes to the messages table. Each message in this relation
references a user (via email) and a ticket (pk).

To get the conversation on a certain ticket the users goes for:
```
GET /tickets/{id}
```

### Tickets status considerations
Since the ticket status is the main benchmark of the whole logic, the staff members should not 
be allowed to call the ticket "resolved" unless a response message from a staff member has been
registered for this tickets.

The ticket's author, on the contrary, may 'close' the ticket at any point of time.

### Flow
If not specified, jwt needed for actions.

##### For common user:
1. Create a user: POST /users (no jwt)
2. Login to set jwt token as cookie: POST /login (no jwt)
3. Create a couple of tickets: POST /tickets
4. List all tickets belonging to the user: GET /tickets

##### For staff user:
1. Create a user  isStaff set to true: POST /users (no jwt, but bearer token NB!)
2. Login to set jwt token as cookie: POST /login (no jwt)
3. List all the tickets: GET /tickets
4. List all the users: GET /users

