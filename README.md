### Users

Users are registere via an API call (POST /users) with the following info in request payload: unique email, at least 8 chars password, and username. By additionally passing true to isStuff boolean field, the user to be registered is claimed to have a stuff status, in which case a bearer token in auth headers expected.

For safety reasons, any other fields added to the request body, incl. isSuperuser, will be ignored.

All users are considered to have common status, unless is_stuff or is_superuser is true in the corresponding 
field of the users table.

### Authorization

To receive a JWT, a post request to /login endpoint expected with email and password specified.