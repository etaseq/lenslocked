/* The industry standard and a good advice to follow is to wait until
   you have clear evidence that you need an INDEX rather than adding
   one too early. It is always possible to come back and add an index
   later but once you do this, it can start to slow things down.*/
/* So DO NOT add an INDEX until you start to see that certain queries
   are slowing your application down and you need to improve the  
   performance of those queries */
CREATE INDEX sessions_token_hash_idx ON SESSIONS(token_hash, user_id, id);

