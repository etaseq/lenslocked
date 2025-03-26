/* These two are exactly the same */
/* When I add the foreign key it means that if I would try to delete
   a user and that user had a session I could not delete him unless I would
   delete the session first.
   But what if a user wants to delete his account? Then I would want him 
   to be able to do this and delete his associated session as well.
   This can be achieved with adding the ON DELETE CASCADE.
*/
CREATE TABLE sessions (
  id SERIAL PRIMARY KEY,
  user_id INT UNIQUE, 
  token_hash TEXT UNIQUE NOT NULL,
  FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
);

CREATE TABLE sessions (
  id SERIAL PRIMARY KEY,
  user_id INT UNIQUE REFERENCES users (id) ON DELETE CASCADE,
  token_hash TEXT UNIQUE NOT NULL
);


/* What to do if I already have a table and I want to add a foreign key to it */
/* sessions_user_id_fkey is just a nice naming convention to adopt. 
   I can use whatever I want */
ALTER TABLE sessions
  ADD CONSTRAINT sessions_user_id_fkey FOREIGN KEY (user_id) REFERENCES users (id);
