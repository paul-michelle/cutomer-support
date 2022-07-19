## Tickets service
A customer support service based of opening a ticket, exchanging messages and, finally, changing the ticket's status.
The info is stored in PostgreSQL. The app itself is written in Go. Apart from postgres driver, jwt library and dotenv file parser, 
only standard libraries used.


### Running the App
Make sure you got [docker](https://www.docker.com/) and [docker-compose](https://docs.docker.com/compose/install/)
installed on your machine.

Clone the repo with *git clone git@github.com:paul-michelle/golang-sql.git*, 
move to the newly created directory and hit *make build* and then *make run* to build and launch the app.
Use *make start* to start the paused containers; *make stop* to stop the containers; 
*make purge* to stop them and remove the data.

Use [Postman](https://www.postman.com/downloads/) - or any other tool of your preference - to execute API calls. 
Please find the *flow* section at the end of this file to quickly test the app. 
For all details - see the corresponding sections below.

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
201 Created || 401 Unauthorized (when creating stuff member) || 405 Method Not Allowed || 400 Bad Request with error details.

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
In case of failure: 401 Unauthorized (if not isStaff or isSuperuser) || 405 Method Not Allowed || 500 Internal Server Error (error reading from the database)

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
200 OK || 405 Method Not Allowed || 404 Not Found || 400 Bad Request || 500 Internal Server Error (jwt token string creation issue, user is asked to retry)

### Tickets (creating and retrieving)
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
Other possible responses: 401 Unauthorized (no or invalid jwt) || 405 Method Not Allowed || 400 Bad Request with error details || 500 Internal Server Error (error when writing to the db.)

To get all the tickets (all user's tickets for common user VS a list of all existing tickets for staff member of superuser):
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
        "author": "postman33@mpost.io", // specified only if it's a request from staff
        "topic": "new instances",
        "status": "pending"
    },
    {
        "id": 2,
        "created_at": "2022-07-16T07:13:58.008489Z",
        "updated_at": "2022-07-16T07:13:58.008489Z",
        "author": "postman33@mpost.io", // specified only if it's a request from staff
        "topic": "new discount",
        "status": "pending"
    },
]
```
Other possible responses: 401 Unauthorized (jwt issues) || 405 Method Not Allowed || 500 Internal Server Error (err when reading from the database)

To retrieve info on a particular ticket the users hits:
```
GET /tickets/{id}
```
In case the ticket exists and the user has got enough permissions to retrieve it:
```
Status 200 OK
{
    "created_at": "2022-07-16T07:14:31.672847Z",
    "updated_at": "2022-07-16T07:14:31.672847Z",
    "topic": "a somewhat complaint",
    "status": "pending"
}
```
Again, as it is the case with listing all the tickets, in case it is a request from a staff memeber, ticket's author also specified
in response body:
```
Status 200 OK
{
    "created_at": "2022-07-16T07:14:31.672847Z",
    "updated_at": "2022-07-16T07:14:31.672847Z",
    "author": "postman33@mpost.io",
    "topic": "third from postman33",
    "status": "pending"
}
```
Other possible responses: 401 Unauthorized (no or invalid jwt) || 405 Method Not Allowed || 404 Not Found.

### Tickets (updating)
A common user can recall the ticket opened by them:
```
PUT/PATCH /tickets/{id}
{
    "status": "canceled"
}
```
Response in case of success is 200 OK. Failures:  401 Unauthorized || 405 Method Not Allowed || 404 Not Found || 400 Bad Request (when user tries to assign an invalid status to the ticket).

Changing the ticket's status, a staff member can choose among "pending", "resolved", "unresolved", e.g.:
```
PUT/PATCH /tickets/{id}
{
    "status": "resolved"
}
```
Response in case of success is 200 OK. Failures:  401 Unauthorized || 405 Method Not Allowed || 404 Not Found || 400 Bad Request (when user tries to assign an invalid status to the ticket, including the "canceled" one - see considerations below).

### Tickets status considerations
A ticket status can be one of the following: "pending" (default), "resolved", "unresolved", "canceled".
Is is only the ticket's author, who can 'close' the ticket via changing its status to "canceled".
The staff members, in their turn, can change the ticket's status to and from "pending", "resolved", "unresolved".

Moreover, the staff are not allowed to call the ticket "resolved" unless at least one response message has been
registered for this tickets.

### Messaging
Each message in this relation references a user (via email) and a ticket (pk). 
A message has got a message type: one of "request", "response", "other". 
When a ticket is first registered, its text goes to the messages table as a message of type "request".

To get the messages on a certain ticket the users goes for: 
```
GET /tickets/{id}/messages.
```
A successful response will be:
```
200 OK
[
    {
        "created_at": "2022-07-17T19:00:44.314775Z",
        "type": "request",
        "text": "a long text with the issue here"
    }
]
```
Other possible responses: 401 Unauthorized (no or invalid jwt) || 405 Method Not Allowed || 404 Not Found.

To add another message to the ticket, the user calls: 
```
POST /tickets/{id}/messages
{
    "text": "another message here..."
}
```
If the message has been sucessfully registered at the back-end, the reponse will be:
```
201 Created
```
Other possible responses: 401 Unauthorized (no or invalid jwt) || 405 Method Not Allowed || 400 Bad Request (missing or invalid payload) || 404 Not Found (ticket not found)

*NB!* If a message is coming from a common user, it is registered as of type "request", while the one from a staff member is
considered to be of type "response".

At the end of the day, the messaging regarding a ticket will be looking like this, when a user calls GET /tickets/{id}/messages:
```
[
    {
        "created_at": "2022-07-17T19:00:44.314775Z",
        "type": "request",
        "text": "I need to know now..."
    },
    {
        "created_at": "2022-07-18T14:14:17.642316Z",
        "type": "request",
        "text": "I cannot wait any longer..."
    },
    {
        "created_at": "2022-07-18T14:18:24.095642Z",
        "type": "request",
        "text": "Yet another message from a user!"
    },
    {
        "created_at": "2022-07-18T14:22:43.930405Z",
        "type": "response",
        "text": "Hi there! Thanks for your patience..."
    }
]
```

### Flow
NB! If not specified, jwt needed for actions.
Let's role-play the communication:

##### For common user:
1. Create a user: POST /users (no jwt)
2. Login to set jwt token as cookie: POST /login (no jwt)
3. Create a couple of tickets: POST /tickets
4. List all tickets belonging to the user: GET /tickets
5. Check messages for the tickets: GET /tickets/{id}/messages
6. Write another message to a ticket: POST /tickets/{id}/messages
7. Check messages again to make sure your recent message is there: GET /tickets/{id}/messages
   
   ...sleep till step 6 is done in the section below...

8. Check messages again to see a response from the admin, if there's a question on whether the ticket can be closed - make a decision.
9.  Alternatively, hit: PUT/PATCH: /tickets/{id}/ Payload {"status": "canceled"}
    
##### For staff user:
1. Create a user  isStaff set to true: POST /users (no jwt, but bearer token NB!)
2. Login to set jwt token as cookie: POST /login (no jwt)
3. List all the existing tickets: GET /tickets
4. Pick up one ticket: GET /tickets/{id}
5. See the messages to the ticket: GET /tickets/{id}/messages
6. If the last message is of type 'request' and there' no mentioning that the ticket can be closed, add a message: POST /tickets/{id}/messages
   
   ...sleep till step 8 is done in the section above...

7. If the user says, the ticket can be closed, hit: PUT/PATCH /tickets/{id}/ Payload {"status": "resolved"}


### Further considerations

Currently, a staff memeber can list all users: *GET /users*.
The list of the users is ordered DESC by the number of tickets a user has opened.

A staff member should be able to hit: *GET /users/{id}* to see the info on this user plus an
embedded array of the tickets. Besides, there should be an option to *PATCH/PUT /users/id* to
change some info on the user, say, user status (which is yet to be implemented).

Having the users details (including the tickets), the staff member can grab a ticket id and
move on to *GET /tickets/{id}* an so on...